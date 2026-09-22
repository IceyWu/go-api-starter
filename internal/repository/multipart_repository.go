package repository

import (
	"github.com/jmoiron/sqlx"
	"go-api-starter/internal/model"
)

var _ MultipartRepositoryInterface = (*MultipartRepository)(nil)

type MultipartRepository struct{ db *sqlx.DB }

func NewMultipartRepository(db *sqlx.DB) *MultipartRepository { return &MultipartRepository{db} }
func (r *MultipartRepository) CreateUpload(v *model.MultipartUpload) error {
	n := modelTime()
	v.CreatedAt = n
	v.UpdatedAt = n
	_, e := r.db.Exec(`INSERT INTO multipart_uploads(upload_id,`+"`key`"+`,md5,file_name,file_size,content_type,total_parts,chunk_size,user_id,status,created_at,updated_at) VALUES(?,?,?,?,?,?,?,?,?,?,?,?)`, v.UploadID, v.Key, v.MD5, v.FileName, v.FileSize, v.ContentType, v.TotalParts, v.ChunkSize, v.UserID, v.Status, n, n)
	return e
}
func (r *MultipartRepository) GetUploadByMD5(m string, u uint) (*model.MultipartUpload, error) {
	var v model.MultipartUpload
	e := r.db.Get(&v, `SELECT * FROM multipart_uploads WHERE md5=? AND user_id=? AND status=? LIMIT 1`, m, u, model.MultipartStatusInitiated)
	return &v, e
}
func (r *MultipartRepository) GetUploadByID(id string) (*model.MultipartUpload, error) {
	var v model.MultipartUpload
	e := r.db.Get(&v, `SELECT * FROM multipart_uploads WHERE upload_id=? AND status=? LIMIT 1`, id, model.MultipartStatusInitiated)
	return &v, e
}
func (r *MultipartRepository) UpdateUploadStatus(id string, s model.MultipartUploadStatus) error {
	_, e := r.db.Exec(`UPDATE multipart_uploads SET status=?,updated_at=? WHERE upload_id=?`, s, modelTime(), id)
	return e
}
func (r *MultipartRepository) DeleteUpload(id string) error {
	_, e := r.db.Exec(`DELETE FROM multipart_uploads WHERE upload_id=?`, id)
	return e
}
func (r *MultipartRepository) SavePart(v *model.UploadedPart) error {
	n := modelTime()
	if v.CreatedAt.IsZero() {
		v.CreatedAt = n
	}
	query := `INSERT INTO uploaded_parts(upload_id,part_number,e_tag,size,created_at) VALUES(?,?,?,?,?) ON CONFLICT(upload_id,part_number) DO UPDATE SET e_tag=excluded.e_tag,size=excluded.size`
	if r.db.DriverName() == "mysql" {
		query = `INSERT INTO uploaded_parts(upload_id,part_number,e_tag,size,created_at) VALUES(?,?,?,?,?) ON DUPLICATE KEY UPDATE e_tag=VALUES(e_tag),size=VALUES(size)`
	}
	_, e := r.db.Exec(query, v.UploadID, v.PartNumber, v.ETag, v.Size, v.CreatedAt)
	return e
}
func (r *MultipartRepository) GetUploadedParts(id string) ([]model.UploadedPart, error) {
	var v []model.UploadedPart
	e := r.db.Select(&v, `SELECT * FROM uploaded_parts WHERE upload_id=? ORDER BY part_number`, id)
	return v, e
}
func (r *MultipartRepository) DeleteParts(id string) error {
	_, e := r.db.Exec(`DELETE FROM uploaded_parts WHERE upload_id=?`, id)
	return e
}
func (r *MultipartRepository) GetUploadedPartNumbers(id string) ([]int, error) {
	var v []int
	e := r.db.Select(&v, `SELECT part_number FROM uploaded_parts WHERE upload_id=? ORDER BY part_number`, id)
	return v, e
}
