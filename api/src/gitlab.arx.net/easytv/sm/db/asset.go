package db

import (
	"database/sql"

	"gitlab.arx.net/easytv/sm"
)

type AssetRepository struct {
	Pool *DatabasePool
}

func (this *AssetRepository) Create(asset *sm.Asset) error {
	stmt, err := this.Pool.Prepare(`
		insert into asset (job_id, path, size, url_param)
		values ($1, $2, $3, $4)
		returning id
	`)

	if err != nil {
		return err
	}

	row := stmt.QueryRow(
		asset.JobID,
		asset.Path,
		asset.Size,
		asset.UrlParam)

	return row.Scan(&asset.ID)
}

func (this *AssetRepository) GetAssetsForJob(job_id int64, assets *[]*sm.Asset) error {
	stmt, err := this.Pool.Prepare(`
		select id, path, size, url_param
		from asset
		where job_id=$1
	`)

	if err != nil {
		return err
	}

	rows, err := stmt.Query(job_id)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		asset := sm.Asset{
			JobID: job_id,
		}
		err = rows.Scan(
			&asset.ID,
			&asset.Path,
			&asset.Size,
			&asset.UrlParam)

		if err != nil {
			return err
		}

		*assets = append(*assets, &asset)
	}

	return nil
}

func (this *AssetRepository) GetAssetsForUser(content_owner_id int64, assets *[]*sm.Asset) error {
	stmt, err := this.Pool.Prepare(`
		select asset.id, asset.path, asset.job_id, asset.size, asset.url_param
		from asset
		inner join job
		on job.id=asset.job_id
		where job.owner_id=$1
	`)

	if err != nil {
		return err
	}

	rows, err := stmt.Query(content_owner_id)

	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		asset := sm.Asset{}
		err = rows.Scan(
			&asset.ID,
			&asset.Path,
			&asset.JobID,
			&asset.Size,
			&asset.UrlParam)

		if err != nil {
			return err
		}

		*assets = append(*assets, &asset)
	}

	return nil
}

func (this *AssetRepository) DeleteAssetsForJob(job_id int64) error {
	stmt, err := this.Pool.Prepare(`
		delete from asset
		where job_id=$1
	`)

	if err != nil {
		return err
	}

	_, err = stmt.Exec(job_id)

	return err
}

func (this *AssetRepository) GetAsset(asset_id int64) (*sm.Asset, error) {
	stmt, err := this.Pool.Prepare(`
		select path, job_id, size, url_param
		from asset
		where id=$1
	`)

	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(asset_id)

	asset := sm.Asset{
		ID: asset_id,
	}

	err = row.Scan(
		&asset.Path,
		&asset.JobID,
		&asset.Size,
		&asset.UrlParam)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (this *AssetRepository) GetAssetByUrlParam(url_param string) (*sm.Asset, error) {
	stmt, err := this.Pool.Prepare(`
		select id, path, job_id, size
		from asset
		where url_param=$1
	`)

	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(url_param)

	asset := sm.Asset{
		UrlParam: url_param,
	}

	err = row.Scan(
		&asset.ID,
		&asset.Path,
		&asset.JobID,
		&asset.Size)

	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &asset, nil
}
