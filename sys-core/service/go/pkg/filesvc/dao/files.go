package dao

import (
	"crypto/sha512"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/genjidb/genji/document"

	utilities "github.com/getcouragenow/sys-share/sys-core/service/config"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

type File struct {
	Id         string `json:"id" genji:"id" coredb:"primary"`
	Binary     []byte `json:"binary" genji:"binary"`
	ShaHash    []byte `json:"sha_hash" genji:"sha_hash"`
	ResourceId string `json:"resource_id" genji:"resource_id"`
	IsDir      bool   `json:"is_dir" genji:"is_dir"`
	CreatedAt  int64  `json:"created_at" genji:"created_at"`
	UpdatedAt  int64  `json:"updated_at" genji:"updated_at"`
}

var (
	filesUniqueIdx = fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS idx_%s_resource_id ON %s(resource_id)", FilesTableName, FilesTableName)
)

func (f File) CreateSQL() []string {
	fields := coredb.GetStructTags(f)
	tbl := coredb.NewTable(FilesTableName, fields, []string{filesUniqueIdx})
	return tbl.CreateTable()
}

func (f *FileDB) UpsertFromUploadRequest(fileByte []byte, id, resourceId string, isDir bool) (*File, error) {
	if len(fileByte) == 0 {
		return nil, fmt.Errorf("empty file")
	}
	qp := coredb.QueryParams{Params: map[string]interface{}{}}
	if id != "" {
		qp.Params["id"] = id
	}
	if resourceId == "" {
		return nil, fmt.Errorf("no resource id")
	}
	qp.Params["resource_id"] = resourceId
	shaHash := sha512.Sum512(fileByte)
	// check existence
	exists, err := f.fileExists(&qp)
	if err != nil || !exists {
		newFile := &File{
			Id:         utilities.NewID(),
			Binary:     fileByte,
			ShaHash:    shaHash[:],
			ResourceId: resourceId,
			CreatedAt:  utilities.CurrentTimestamp(),
			UpdatedAt:  utilities.CurrentTimestamp(),
		}
		if isDir {
			newFile.IsDir = true
		}
		filterParam, err := coredb.AnyToQueryParam(newFile, true)
		if err != nil {
			return nil, err
		}
		columns, values := filterParam.ColumnsAndValues()
		if len(columns) != len(values) {
			return nil, fmt.Errorf("error: length mismatch: cols: %d, vals: %d", len(columns), len(values))
		}
		stmt, args, err := sq.Insert(FilesTableName).
			Columns(columns...).
			Values(values...).
			ToSql()
		if err = f.db.Exec(stmt, args...); err != nil {
			return nil, err
		}
		qp.Params["id"] = newFile.Id
	} else {
		file, err := f.Get(&qp)
		if err != nil {
			return nil, err
		}
		if string(file.ShaHash) == string(shaHash[:]) {
			return file, nil
		}
		updFile := &File{
			Id:         file.Id,
			Binary:     fileByte,
			ShaHash:    shaHash[:],
			ResourceId: file.ResourceId,
			CreatedAt:  file.CreatedAt,
			UpdatedAt:  utilities.CurrentTimestamp(),
		}
		if isDir {
			updFile.IsDir = true
		}
		filterParam, err := coredb.AnyToQueryParam(updFile, true)
		if err != nil {
			return nil, err
		}
		delete(filterParam.Params, "id")
		delete(filterParam.Params, "resource_id")
		stmt, args, err := sq.Update(FilesTableName).SetMap(filterParam.Params).Where(sq.Eq{"id": file.Id}).ToSql()
		if err != nil {
			return nil, err
		}
		if err = f.db.Exec(stmt, args...); err != nil {
			return nil, err
		}
	}
	nf, err := f.Get(&qp)
	if err != nil {
		return nil, err
	}
	return nf, nil
}

func (f *FileDB) fileExists(query *coredb.QueryParams) (bool, error) {
	baseStmt := sq.Select("COUNT(id)").From(FilesTableName)
	for k, v := range query.Params {
		baseStmt = baseStmt.Where(sq.Eq{k: v})
	}
	stmt, args, err := baseStmt.ToSql()
	if err != nil {
		return false, err
	}
	doc, err := f.db.QueryOne(stmt, args...)
	if err != nil {
		if err.Error() == "document not found" {
			return false, nil
		}
		return false, err
	}
	c, err := doc.Doc.GetByField("COUNT(id)")
	if err != nil {
		return false, err
	}
	if c.V.(int64) == 0 {
		return false, nil
	}
	return true, nil
}

func (f *FileDB) Get(query *coredb.QueryParams) (*File, error) {
	baseStmt := sq.Select(f.fileColumns).From(FilesTableName)
	for k, v := range query.Params {
		baseStmt = baseStmt.Where(sq.Eq{k: v})
	}
	stmt, args, err := baseStmt.ToSql()
	if err != nil {
		return nil, err
	}
	var nf File
	doc, err := f.db.QueryOne(stmt, args...)
	if err != nil {
		return nil, err
	}
	if err = document.StructScan(doc.Doc, &nf); err != nil {
		return nil, err
	}
	return &nf, nil
}
