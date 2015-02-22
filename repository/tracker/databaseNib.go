package tracker

import (
	"os"

	"github.com/jinzhu/gorm"

	_ "github.com/mattn/go-sqlite3"
)

type NIBLookup struct {
	NIBID string `sql:"size:256"`
	Path  string `sql:"size:1024"`
}

func NewDatabaseNIBTracker(dbLocation string) (NIBTracker, error) {
	nibTracker := DatabaseNIBTracker{
		dbLocation: dbLocation,
	}
	_, statErr := os.Stat(dbLocation)

	db, err := gorm.Open("sqlite3", nibTracker.dbLocation)
	nibTracker.db = db
	if err == nil && os.IsNotExist(statErr) {
		err = nibTracker.createDb()
	}

	return nibTracker, err
}

type DatabaseNIBTracker struct {
	dbLocation string
	db         gorm.DB
}

// createDb initializes the tables in the database structure.
func (d *DatabaseNIBTracker) createDb() error {
	db := d.db.CreateTable(&NIBLookup{})
	return db.Error
}

// Add registers the given nibID for the given path.
func (d *DatabaseNIBTracker) Add(path string, nibID string) error {
	db := d.db.Create(&NIBLookup{
		NIBID: nibID,
		Path:  path,
	})
	return db.Error
}

// whereFor returns a where statement for the
func (d *DatabaseNIBTracker) whereFor(path string) *gorm.DB {
	return d.db.Where(map[string]interface{}{"path": path})
}

// lookupToNIB converts the lookup nib to a search response.
func (d *DatabaseNIBTracker) lookupToNIB(nibLookup *NIBLookup) *NIBSearchResponse {
	return &NIBSearchResponse{
		NIBID:          nibLookup.NIBID,
		Path:           nibLookup.Path,
		repositoryPath: "",
	}
}

// Get returns the nibID for the given path.
func (d *DatabaseNIBTracker) Get(path string) (*NIBSearchResponse, error) {
	stmt := d.whereFor(path)
	data := &NIBLookup{}
	db := stmt.First(data)
	if db.Error != nil {
		return nil, db.Error
	}

	return d.lookupToNIB(data), nil
}

// SearchPrefix returns all nibIDs with the given path.
// The map being returned has the paths
func (d *DatabaseNIBTracker) SearchPrefix(prefix string) ([]*NIBSearchResponse, error) {
	var resp []NIBLookup
	db := d.db.Where("path LIKE ?", prefix+"%").Find(&resp)

	searchResponse := []*NIBSearchResponse{}
	for _, item := range resp {
		searchResponse = append(searchResponse, d.lookupToNIB(item))
	}

	return searchResponse, db.Error
}

// Remove removes the given path from being tracked.
func (d *DatabaseNIBTracker) Remove(path string) error {
	db := d.whereFor(path).Delete(NIBLookup{})
	return db.Error
}
