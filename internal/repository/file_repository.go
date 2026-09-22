package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"go-api-starter/internal/model"
)

var ErrFileNotFound = errors.New("file not found")
var _ FileRepositoryInterface = (*FileRepository)(nil)

type FileRepository struct{ db *sqlx.DB }

func NewFileRepository(db *sqlx.DB) *FileRepository { return &FileRepository{db} }

const fileColumns = "id,uid,user_id,name,type,file_md5,size,`key`,extension,width,height,blurhash,arthash,arthash_codec,lng,lat,country,country_code,province,city,district,address,altitude,taken_at,device_make,device_model,lens_model,f_number,exposure_time,iso,focal_length,exif_raw,duration,codec,bitrate,frame_rate,video_metadata,transcoding_task_id,is_private,created_at,updated_at"
const fileKeyColumn = "`key`"

func (r *FileRepository) Create(c context.Context, v *model.File) error {
	v.PrepareForCreate()
	n := modelTime()
	v.CreatedAt = n
	v.UpdatedAt = n
	x, e := r.db.ExecContext(c, `INSERT INTO files(uid,user_id,name,type,file_md5,size,`+fileKeyColumn+`,extension,width,height,blurhash,arthash,arthash_codec,lng,lat,country,country_code,province,city,district,address,altitude,taken_at,device_make,device_model,lens_model,f_number,exposure_time,iso,focal_length,exif_raw,duration,codec,bitrate,frame_rate,video_metadata,transcoding_task_id,is_private,created_at,updated_at) VALUES(
?,?,?,?,?,?,?,?,?, ?,
?,?,?,?,?,?,?,?,?, ?,
?,?,?,?,?,?,?,?,?, ?,
?,?,?,?,?,?,?,?,?, ?,
?)`, v.UID, v.UserID, v.Name, v.Type, v.FileMd5, v.Size, v.Key, v.Extension, v.Width, v.Height, v.Blurhash, v.Arthash, v.ArthashCodec, v.Lng, v.Lat, v.Country, v.CountryCode, v.Province, v.City, v.District, v.Address, v.Altitude, v.TakenAt, v.DeviceMake, v.DeviceModel, v.LensModel, v.FNumber, v.ExposureTime, v.ISO, v.FocalLength, v.ExifRaw, v.Duration, v.Codec, v.Bitrate, v.FrameRate, v.VideoMetadata, v.TranscodingTaskID, v.IsPrivate, n, n)
	if e == nil {
		if id, z := x.LastInsertId(); z == nil {
			v.ID = uint(id)
		}
	}
	if e == nil {
		v.PrepareForResponse()
	}
	return e
}
func (r *FileRepository) FindByID(c context.Context, id uint) (*model.File, error) {
	return r.find(c, `id=?`, id)
}
func (r *FileRepository) FindByMD5(c context.Context, v string) (*model.File, error) {
	return r.find(c, `file_md5=?`, v)
}
func (r *FileRepository) FindByUID(c context.Context, v string) (*model.File, error) {
	return r.find(c, `uid=?`, v)
}
func (r *FileRepository) find(c context.Context, w string, a any) (*model.File, error) {
	var v model.File
	e := r.db.GetContext(c, &v, `SELECT `+fileColumns+` FROM files WHERE `+w+` LIMIT 1`, a)
	if errors.Is(e, sql.ErrNoRows) {
		return nil, ErrFileNotFound
	}
	if e == nil {
		v.PrepareForResponse()
	}
	return &v, e
}
func (r *FileRepository) Update(c context.Context, v *model.File) error {
	v.UpdatedAt = modelTime()
	_, e := r.db.ExecContext(c, `UPDATE files SET name=?,type=?,file_md5=?,size=?,`+fileKeyColumn+`=?,extension=?,width=?,height=?,blurhash=?,arthash=?,arthash_codec=?,lng=?,lat=?,country=?,country_code=?,province=?,city=?,district=?,address=?,altitude=?,taken_at=?,device_make=?,device_model=?,lens_model=?,f_number=?,exposure_time=?,iso=?,focal_length=?,exif_raw=?,duration=?,codec=?,bitrate=?,frame_rate=?,video_metadata=?,transcoding_task_id=?,is_private=?,updated_at=? WHERE id=?`, v.Name, v.Type, v.FileMd5, v.Size, v.Key, v.Extension, v.Width, v.Height, v.Blurhash, v.Arthash, v.ArthashCodec, v.Lng, v.Lat, v.Country, v.CountryCode, v.Province, v.City, v.District, v.Address, v.Altitude, v.TakenAt, v.DeviceMake, v.DeviceModel, v.LensModel, v.FNumber, v.ExposureTime, v.ISO, v.FocalLength, v.ExifRaw, v.Duration, v.Codec, v.Bitrate, v.FrameRate, v.VideoMetadata, v.TranscodingTaskID, v.IsPrivate, v.UpdatedAt, v.ID)
	return e
}
func (r *FileRepository) Delete(c context.Context, id uint) error {
	x, e := r.db.ExecContext(c, `DELETE FROM files WHERE id=?`, id)
	if e != nil {
		return e
	}
	n, _ := x.RowsAffected()
	if n == 0 {
		return ErrFileNotFound
	}
	return nil
}
func (r *FileRepository) List(c context.Context, f model.FileFilter, off, limit int, sort string) ([]model.File, int64, error) {
	where := ` WHERE 1=1`
	args := []any{}
	if f.UserID != nil {
		where += ` AND user_id=?`
		args = append(args, *f.UserID)
	}
	if f.Type != nil {
		where += ` AND type=?`
		args = append(args, *f.Type)
	}
	if f.IsPrivate != nil {
		where += ` AND is_private=?`
		args = append(args, *f.IsPrivate)
	}
	var total int64
	if e := r.db.GetContext(c, &total, `SELECT COUNT(*) FROM files`+where, args...); e != nil {
		return nil, 0, e
	}
	if sort == `` || !safeSort(sort) {
		sort = `created_at DESC`
	}
	var v []model.File
	e := r.db.SelectContext(c, &v, fmt.Sprintf(`SELECT %s FROM files%s ORDER BY %s LIMIT ? OFFSET ?`, fileColumns, where, sort), append(args, limit, off)...)
	for i := range v {
		v[i].PrepareForResponse()
	}
	return v, total, e
}
