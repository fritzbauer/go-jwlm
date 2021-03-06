package model

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/tj/assert"
)

var exampleManifest = &manifest{
	CreationDate: time.Now().Format("2006-01-02"),
	UserDataBackup: userDataBackup{
		LastModifiedDate: time.Now().Format("2006-01-02T15:04:05-07:00"),
		Hash:             "f57aabf8f375aa5469e3aea2292f89d2f624b8b2d70e0e0688f9ffbd44f0cf2b",
		DatabaseName:     "user_data.db",
		SchemaVersion:    8,
		DeviceName:       "go-jwlm",
	},
	Name:    "test",
	Type:    0,
	Version: 1,
}

func Test_manifest_importManifest(t *testing.T) {
	path := filepath.Join("testdata", "manifest_correct.json")

	mfst := &manifest{}
	assert.NoError(t, mfst.importManifest(path))

	expectedMfst := &manifest{
		CreationDate: "2020-04-11",
		UserDataBackup: userDataBackup{
			LastModifiedDate: "2020-04-09T05:47:26+02:00",
			Hash:             "d87a67028133cc4de5536affe1b072841def95899b7f7450a5622112b4b5e63f",
			DatabaseName:     "user_data.db",
			SchemaVersion:    8,
			DeviceName:       "iPhone",
		},
		Name:    "UserDataBackup_2020-04-11_iPhone",
		Type:    0,
		Version: 1,
	}
	assert.Equal(t, expectedMfst, mfst)

	assert.Error(t, mfst.importManifest("nonexistentpath"))
}

func Test_validateManifest(t *testing.T) {
	path := filepath.Join("testdata", "manifest_correct.json")

	mfst := manifest{}
	assert.NoError(t, mfst.importManifest(path))
	assert.NoError(t, mfst.validateManifest())

	path = filepath.Join("testdata", "manifest_outdated.json")
	mfst = manifest{}
	assert.NoError(t, mfst.importManifest(path))
	assert.Error(t, mfst.validateManifest())
}

func Test_generateManifest(t *testing.T) {
	dbPath := filepath.Join("testdata", "user_data.db")

	mfst, err := generateManifest("test", dbPath)
	exampleManifest.UserDataBackup.LastModifiedDate = time.Now().Format("2006-01-02T15:04:05-07:00") // Could have changed in the last second..
	assert.NoError(t, err)
	assert.Equal(t, exampleManifest, mfst)

	_, err = generateManifest("test", "nonexistent.db")
	assert.Error(t, err)
}

func Test_exportManifest(t *testing.T) {
	tmp, err := ioutil.TempDir("", "go-jwlm")
	assert.NoError(t, err)
	//defer os.RemoveAll(tmp)

	path := filepath.Join(tmp, "test_manifest.json")
	fmt.Println(path)
	err = exampleManifest.exportManifest(path)
	assert.NoError(t, err)
	assert.FileExists(t, path)

	otherMfst := &manifest{}
	err = otherMfst.importManifest(path)
	assert.NoError(t, err)
	assert.Equal(t, exampleManifest, otherMfst)

}
