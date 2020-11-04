package dao

import (
	"crypto/sha512"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/getcouragenow/sys/sys-core/service/go/pkg/coredb"
)

type File struct {
	Id        string `json:"id" genji:"id" coredb:"primary"`
	Binary    []byte `json:"binary" genji:"binary"`
	Sum       string `json:"sum" genji:"sum"`
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
	// check existence
	exists, err := f.fileExists(&qp)
	if err != nil || !exists {
		sum := sha512.Sum512(fileByte)
		newFile := &File{
			Id:        coredb.NewID(),
			Binary:    fileByte,
			Sum:       string(sum[:]),
			ForeignId: foreignId,
			CreatedAt: coredb.CurrentTimestamp(),
			UpdatedAt: coredb.CurrentTimestamp(),
		}
		stmt := fmt.Sprintf("INSERT INTO %s(%s) VALUES(?, ?, ?, ?, ?, ?)", FilesTableName, f.fileColumns)
		args := []interface{}{newFile.Id, newFile.Binary, newFile.Sum, newFile.ForeignId, newFile.CreatedAt, newFile.UpdatedAt}
		if err = f.db.Exec(stmt, args...); err != nil {
			return nil, err
		}
		qp.Params["id"] = newFile.Id

	} else {
		file, err := f.Get(&qp)
		if err != nil {
			return nil, err
		}
		sum := sha512.Sum512(fileByte)
		if file.Sum == string(sum[:]) {
			return file, nil
		}
		file.UpdatedAt = coredb.CurrentTimestamp()
		file.Sum = string(sum[:])
		file.Binary = fileByte
		filterParam, err := coredb.AnyToQueryParam(file, true)
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
	if c.V.(int) != 0 {
		return false, nil
	}
	return true, nil
}

func (f *FileDB) Get(query *coredb.QueryParams) (*File, error) {
	baseStmt := sq.Select().From(FilesTableName)
	for k, v := range query.Params {
		baseStmt = baseStmt.Where(sq.Eq{k: v})
	}
	stmt, args, err := baseStmt.ToSql()
	if err != nil {
		return nil, err
	}
	var nf *File
	doc, err := f.db.QueryOne(stmt, args...)
	if err != nil {
		return nil, err
	}
	if err = doc.StructScan(nf); err != nil {
		return nil, err
	}
	return nf, nil
}
