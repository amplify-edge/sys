package repo

import (
	"bytes"
	"encoding/binary"
	corepkg "github.com/getcouragenow/sys-share/sys-core/service/go/rpc/v2"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/filesvc/dao"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"math"
)

const (
	maxFileSize       = 4 << 20 // max file upload is 4 MB
	downloadChunkSize = 1 << 20 / 2
)

type SysFileRepo struct {
	store *dao.FileDB
	log   *logrus.Entry
}

func NewSysFileRepo(db *coredb.CoreDB, log *logrus.Entry) (*SysFileRepo, error) {
	store, err := dao.NewFileDB(db, log)
	if err != nil {
		return nil, err
	}
	return &SysFileRepo{
		store: store,
		log:   log,
	}, nil
}

func (s *SysFileRepo) Upload(stream corepkg.FileService_UploadServer) error {
	req, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "cannot upload file: %v", err)
	}
	var foreignId string
	if req.GetFileInfo().GetSysAccountId() != "" {
		foreignId = req.GetFileInfo().GetSysAccountId()
	}
	if req.GetFileInfo().GetSysAccountOrgId() != "" {
		foreignId = req.GetFileInfo().GetSysAccountOrgId()
	}
	if req.GetFileInfo().GetSysAccountProjectId() != "" {
		foreignId = req.GetFileInfo().SysAccountProjectId
	}
	isDir := req.GetFileInfo().GetIsDir()
	fileBuf := bytes.Buffer{}
	fileSize := 0
	for {
		req, err = stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return status.Errorf(codes.Unknown, "cannot receive stream chunk: %v", err)
		}
		chunk := req.GetChunk()
		chunkSize := len(chunk)
		fileSize += chunkSize
		if fileSize > maxFileSize {
			return status.Error(codes.Aborted, "file size exceeds maximum file size, aborting")
		}
		_, err = fileBuf.Write(chunk)
		if err != nil {
			return status.Errorf(codes.Internal, "cannot write file data to buffer: %v", err)
		}
	}

	f, err := s.store.UpsertFromUploadRequest(fileBuf.Bytes(), "", foreignId, isDir)
	if err != nil {
		return status.Errorf(codes.Internal, "cannot save file to db: %v", err)
	}
	resp := &corepkg.FileUploadResponse{
		Success:  true,
		Id:       f.Id,
		ErrorMsg: "",
	}
	if err = stream.SendAndClose(resp); err != nil {
		return status.Errorf(codes.Internal, "cannot encode upload resp: %v", err)
	}
	s.log.Debugf("Saved file, id: %s, sum: %v, size: %d", f.Id, f.Sum, fileSize)
	return nil
}

func (s *SysFileRepo) Download(req *corepkg.FileDownloadRequest, stream corepkg.FileService_DownloadServer) error {
	f, err := s.store.Get(&coredb.QueryParams{Params: map[string]interface{}{"id": req.GetId()}})
	if err != nil {
		return status.Errorf(codes.NotFound, "cannot found an object of id: %s => %v", req.GetId(), err)
	}
	fileSize := binary.Size(f.Binary)
	filePartsCount := fileSize/downloadChunkSize + 1
	offset := 0
	for i := 0; i < filePartsCount; i++ {
		partSize := int(math.Min(downloadChunkSize, float64(fileSize-(i*downloadChunkSize))))
		s.log.Debugf("Sending partsize of size %d to client", partSize)
		chunk := f.Binary[offset:(offset + partSize)]
		resp := &corepkg.FileDownloadResponse{Chunk: chunk, TotalSize: int64(fileSize)}
		if f.IsDir {
			resp.IsCompressed = true
		}
		if err = stream.Send(resp); err != nil {
			return status.Errorf(codes.Internal, "cannot send chunk to be downloaded: %v", err)
		}
	}
	return nil
}
