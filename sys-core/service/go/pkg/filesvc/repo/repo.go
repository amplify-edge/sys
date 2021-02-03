package repo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/amplify-cms/sys-share/sys-core/service/logging"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"math"

	"github.com/amplify-cms/sys-share/sys-core/service/go/pkg/filehelper"
	corepkg "github.com/amplify-cms/sys-share/sys-core/service/go/rpc/v2"
	"github.com/amplify-cms/sys/sys-core/service/go/pkg/coredb"
	"github.com/amplify-cms/sys/sys-core/service/go/pkg/filesvc/dao"
)

const (
	maxFileSize       = 4 << 20 // max file upload is 4 MB
	downloadChunkSize = 1 << 20 / 2
)

type SysFileRepo struct {
	store *dao.FileDB
	log   logging.Logger
}

func NewSysFileRepo(db *coredb.CoreDB, log logging.Logger) (*SysFileRepo, error) {
	store, err := dao.NewFileDB(db, log)
	if err != nil {
		return nil, err
	}
	return &SysFileRepo{
		store: store,
		log:   log,
	}, nil
}

func (s *SysFileRepo) sharedUpload(content []byte, resourceId string, isDir bool) (*corepkg.FileUploadResponse, error) {
	f, err := s.store.UpsertFromUploadRequest(content, "", resourceId, isDir)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot save file to db: %v", err)
	}
	resp := &corepkg.FileUploadResponse{
		Success:    true,
		Id:         f.Id,
		ResourceId: f.ResourceId,
		ErrorMsg:   "",
	}
	return resp, nil
}

// UploadFile for v2, Upload for v3
func (s *SysFileRepo) UploadFile(filepath string, content []byte) (*corepkg.FileUploadResponse, error) {
	var finfo *corepkg.FileInfo
	var err error
	fileContent := content
	if filepath != "" && (content == nil && len(content) == 0) {
		finfo, fileContent, err = filehelper.ReadFileFromPath(filepath)
		if err != nil {
			return nil, err
		}
	} else if filepath == "" && (content != nil && len(content) != 0) {
		finfo, fileContent, err = filehelper.ReadFileFromBytes(filepath, content)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("filepath: %s, content: %v", filepath, content)
	}
	s.log.Debugf("")
	return s.sharedUpload(fileContent, finfo.GetResourceId(), finfo.GetIsDir())
}

func (s *SysFileRepo) Upload(stream corepkg.FileService_UploadServer) error {
	req, err := stream.Recv()
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "cannot upload file: %v", err)
	}
	resourceId := req.FileInfo.ResourceId
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

	resp, err := s.sharedUpload(fileBuf.Bytes(), resourceId, isDir)
	if err = stream.SendAndClose(resp); err != nil {
		return status.Errorf(codes.Internal, "cannot encode upload resp: %v", err)
	}
	s.log.Debugf("Saved file, id: %s", resp.Id)
	return nil
}

func (s *SysFileRepo) DownloadFile(fileId, resourceId string) (*dao.File, error) {
	params := map[string]interface{}{}
	if fileId != "" {
		params["id"] = fileId
	}
	if resourceId != "" {
		params["resource_id"] = resourceId
	}
	f, err := s.store.Get(&coredb.QueryParams{Params: params})
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "cannot found an object of id: %s / resource_id: %s => %v", fileId, resourceId, err)
	}
	return f, nil
}

func (s *SysFileRepo) Download(req *corepkg.FileDownloadRequest, stream corepkg.FileService_DownloadServer) error {
	f, err := s.DownloadFile(req.GetId(), "")
	if err != nil {
		return err
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
