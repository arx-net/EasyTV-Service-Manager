package sm

import "mime/multipart"

type Asset struct {
	ID       int64
	JobID    int64
	UrlParam string
	Path     string
	Size     int64
}

type AssetRepository interface {
	Create(asset *Asset) error

	GetAssetsForJob(job_id int64, assets *[]*Asset) error

	GetAssetsForUser(content_owner_id int64, assets *[]*Asset) error

	DeleteAssetsForJob(job_id int64) error

	GetAsset(asset_id int64) (*Asset, error)

	GetAssetByUrlParam(url_param string) (*Asset, error)
}

type AssetService interface {
	CreateAsset(step_id int64,
		module *Module,
		file multipart.File,
		filename string,
		filesize int64) (*Asset, error)

	GC() error
}
