package tracker

import (
	"errors"
	"os"
	"strings"

	"github.com/jinzhu/gorm"

	// Needed for sqlite gorm support.
	_ "github.com/mattn/go-sqlite3"
)

// NIBLookup struct is used to represent entries in the database.
type NIBLookup struct {
	ID    int64
	NIBID string `sql:"size:256;unique" gorm:"column:nib_id"`
	Path  string `sql:"size:4096;unique"`
}

// TableName returns the name of the SQLite NIB table.
func (n NIBLookup) TableName() string {
	return "nib_lookups"
}

// NewDatabaseNIBTracker initializes a new object which uses a database
// to track NIB changes and implements the NIBTracker repository.
func NewDatabaseNIBTracker(dbLocation string, repositoryPath string) (NIBTracker, error) {
	nibTracker := &DatabaseNIBTracker{
		dbLocation:     dbLocation,
		repositoryPath: repositoryPath,
	}
	_, statErr := os.Stat(dbLocation)

	db, err := gorm.Open("sqlite3", nibTracker.dbLocation)
	nibTracker.db = &db
	if err == nil && os.IsNotExist(statErr) {
		err = nibTracker.createDb()
	}

	return nibTracker, err
}

// DatabaseNIBTracker implements the NIBTracker interface and utilizes
// a sqlite database backend for persistence.
type DatabaseNIBTracker struct {
	dbLocation     string
	db             *gorm.DB
	repositoryPath string
}

// createDb initializes the tables in the database structure.
func (d *DatabaseNIBTracker) createDb() error {
	db := d.db.CreateTable(&NIBLookup{})
	return db.Error
}

// Add registers the given nibID for the given path.
func (d *DatabaseNIBTracker) Add(path string, nibID string) error {
	if len(path) > MaxPathSize {
		return errors.New("Path longer than maximal allowed path.")
	}
	tx := d.db.Begin()
	res, err := d.getLookup(path, tx)

	var db *gorm.DB
	if err == nil && res != nil {
		res.NIBID = nibID
		db = tx.Save(res)
	} else {
		res = &NIBLookup{
			NIBID: nibID,
			Path:  path,
		}
		db = tx.Create(res)
	}

	tx.Commit()
	return db.Error
}

// whereFor returns a where statement which requests entries from the database
// for the passed path.
func (d *DatabaseNIBTracker) whereFor(path string, db *gorm.DB) *gorm.DB {
	return db.Where(map[string]interface{}{"path": path})
}

// lookupToNIB converts the lookup nib to a search response.
func (d *DatabaseNIBTracker) lookupToNIB(nibLookup *NIBLookup) *NIBSearchResponse {
	return &NIBSearchResponse{
		NIBID:          nibLookup.NIBID,
		Path:           nibLookup.Path,
		repositoryPath: d.repositoryPath,
	}
}

// get returns the database object for the given path.
func (d *DatabaseNIBTracker) getLookup(path string, db *gorm.DB) (*NIBLookup, error) {
	stmt := d.whereFor(path, db)
	data := &NIBLookup{}
	res := stmt.First(data)
	if res.Error != nil {
		return nil, res.Error
	}
	return data, nil
}

// Get returns the nibID for the given path.
func (d *DatabaseNIBTracker) Get(path string) (*NIBSearchResponse, error) {
	data, err := d.getLookup(path, d.db)
	if err != nil {
		return nil, err
	}

	return d.lookupToNIB(data), err
}

// SearchPrefix returns all nibIDs with the given path.
// The map being returned has the paths
func (d *DatabaseNIBTracker) SearchPrefix(prefix string) ([]*NIBSearchResponse, error) {
	var resp []NIBLookup

	prefix = strings.TrimSuffix(prefix, "/")
	directoryPrefix := prefix + "/"
	db := d.db.Where("path LIKE ? or path = ?", directoryPrefix+"%", prefix).Find(&resp)

	searchResponse := []*NIBSearchResponse{}
	for _, item := range resp {
		searchResponse = append(searchResponse, d.lookupToNIB(&item))
	}

	return searchResponse, db.Error
}

// Remove removes the given path from being tracked.
func (d *DatabaseNIBTracker) Remove(path string) error {
	tx := d.db.Begin()
	db := d.whereFor(path, tx).Delete(NIBLookup{})
	if db.Error != nil {
		tx.Rollback()
	} else if db.Error == nil && db.RowsAffected < 1 {
		tx.Rollback()
		return errors.New("Entry not found")
	} else {
		tx.Commit()
	}
	return db.Error
}
