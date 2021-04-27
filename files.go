package main

import (
	"fmt"
	"io/fs"
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"time"
)

type CamerasDir struct {
	fs.FS
	Cameras map[int]CameraDir
}

type CameraDir struct {
	fs.DirEntry
	Num string
	Videos map[time.Time]VideoFile
}

type VideoFile struct {
	fs.DirEntry
    Time time.Time
}

func GetCams(rootdir string) *CamerasDir {
	cams := CamerasDir{
		FS: os.DirFS("/"), //rootdir),
		Cameras: make(map[int]CameraDir, 10),
	}
	vidre := regexp.MustCompile("(/?.*/|)(pik_(\\d\\d\\d)_\\d\\d\\d\\d-\\d\\d-\\d\\d_\\d\\d-\\d\\d-\\d\\d)\\.(.*?)")
	files, err := fs.Glob(cams, "*/MP4/*.mp4")
	if err != nil {
		fmt.Errorf("reading filelist error: %v", err)
		return nil
	}
	for _, fname := range files {
		m := vidre.FindAllStringSubmatch(fname, -1)
		if m != nil {



		}
	}

	var dirofs int
	if len(rootdir) > 0 && rootdir[0:1] == "/" { dirofs = 1 }
	camsdir, _ := fs.ReadDir(cams.FS, rootdir[dirofs:])

	camre := regexp.MustCompile("^\\d\\d\\d$")
	for _, cam := range camsdir {
		if cam.IsDir() && camre.Match([]byte(cam.Name())) {
			num, _ := strconv.Atoi(cam.Name())
			cams.Cameras[num] = CameraDir{
				DirEntry: cam,
				Num: cam.Name(),
				Videos: make(map[time.Time]VideoFile),
			}
		}

	}
    return &cams
}

func (dir *CameraDir) Update(filterTime [2]time.Time) {

	camsdir, _ := fs.ReadDir(dir, "MP4")

	vidre := regexp.MustCompile("(/?.*/|)(pik_(\\d\\d\\d)_\\d\\d\\d\\d-\\d\\d-\\d\\d_\\d\\d-\\d\\d-\\d\\d)\\.(.*?)")
	for _, vid := range camsdir {
		m := vidre.FindAllStringSubmatchIndex(vid, -1)
		if m == nil {
			bot.Send(msg.Sender, "Could not parse filename from args")
			return
		}
		if cam.IsDir() && camre.Match([]byte(cam.Name())) {
			num, _ := strconv.Atoi(cam.Name())
			cams.Cameras[num] = CameraDir{
				DirEntry: cam,
				Num: cam.Name(),
				Videos: make(map[time.Time]VideoFile),
			}
		}

	}
}

type FileList struct {
	JSONFile string
	Data *FileListData
	Diff [3]*FileListData // Inserted, Updated, Deleted
	func Update(data *FileListData) error
	func Load() error
	func Save() error
}

type FileListData struct {
	LastTime time.Time `json:"last_time"`
	FileList []FileListItem `json:"files"`
}

type FileListItem struct {
	Time time.Time `json:"time"`
	File string `json:"file"`
	Active bool `json:"active"`
}

func NewFileList(jsonFile string) *FileList {
	var fileList FileList
	if len(jsonFile) == 0 {
		fileList.jsonFile = "filelist.json"
	} else {
		fileList.JSONFile = jsonFile
	}
	fileList.Data = &FileListData{}
	return &fileList
}

func (fileList *FileList) Load() error {
	json, err := ioutil.ReadFile(fileList.JSONFile)
	if err != nil {

		return err
	}

	data := &FileListData{}
	err = json.Unmarshal([]byte(json), data)
	if err != nil {

		return err
	}

	return fileList.Update(data)
}

func (fileList *FileList) SaveFileList(filename string) *FileList {
	var json string
	err = json.Marshal([]byte(fileContent), filedata)
	if len(filename) == 0 {
		filename = "filelist.json"
	}
	fileContent, err := ioutil.ReadFile(filename)
	if err != nil {
		return &FileList{}
	}
	if err == nil {
		return filedata
	}
	return nil
}

func (fileList *FileList) Update(data *FileListData) error {

	return nil
}