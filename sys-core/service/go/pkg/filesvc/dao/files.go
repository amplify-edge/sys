package dao

import (
	"crypto/sha512"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/genjidb/genji/document"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

type File struct {
	Id        string `json:"id" genji:"id" coredb:"primary"`
	Binary    []byte `json:"binary" genji:"binary"`
	Sum       []byte `json:"sum" genji:"sum"`
	ForeignId string `json:"foreignId" genji:"foreign_id"`
	CreatedAt int64  `json:"createdAt" genji:"created_at"`
	UpdatedAt int64  `json:"updatedAt" genji:"updated_at"`
}

var (
	filesUniqueIdx = fmt.Sprintf("CREATE UNIQUE INDEX IF NOT EXISTS idx_%s_foreign_id ON %s(foreign_id)", FilesTableName, FilesTableName)
)

func (f File) CreateSQL() []string {
	fields := coredb.GetStructTags(f)
	tbl := coredb.NewTable(FilesTableName, fields, []string{filesUniqueIdx})
	return tbl.CreateTable()
}

func (f *FileDB) UpsertFromUploadRequest(fileByte []byte, id, foreignId string) (*File, error) {
	if len(fileByte) == 0 {
		return nil, fmt.Errorf("empty file")
	}
	qp := coredb.QueryParams{Params: map[string]interface{}{}}
	if id != "" {
		qp.Params["id"] = id
	}
	if foreignId == "" {
		return nil, fmt.Errorf("no foreign id")
	}
	qp.Params["foreign_id"] = foreignId
	sum := sha512.Sum512(fileByte)
	// check existence
	exists, err := f.fileExists(&qp)
	if err != nil || !exists {
		newFile := &File{
			Id:        coredb.NewID(),
			Binary:    fileByte,
			Sum:       sum[:],
			ForeignId: foreignId,
			CreatedAt: coredb.CurrentTimestamp(),
			UpdatedAt: coredb.CurrentTimestamp(),
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
		if string(file.Sum) == string(sum[:]) {
			return file, nil
		}
		updFile := &File{
			Id:        file.Id,
			Binary:    fileByte,
			Sum:       sum[:],
			ForeignId: file.ForeignId,
			CreatedAt: file.CreatedAt,
			UpdatedAt: coredb.CurrentTimestamp(),
		}
		filterParam, err := coredb.AnyToQueryParam(updFile, true)
		if err != nil {
			return nil, err
		}
		delete(filterParam.Params, "id")
		delete(filterParam.Params, "foreign_id")
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
