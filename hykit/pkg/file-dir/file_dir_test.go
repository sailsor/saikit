package filedir

import (
	"os"
	"testing"

	"code.jshyjdtech.com/godev/hykit/log"
	"github.com/stretchr/testify/assert"
)

const (
	file = "./test/esim.txt"

	dir = "./test_dir"
)

func TestIsEmptyDir(t *testing.T) {
	empty, err := IsEmptyDir(".")
	if err != nil {
		t.Error(err.Error())
	}

	if empty == true {
		t.Error("结果错误，目录不为空")
	}
}

func TestIsEmptyDir2(t *testing.T) {
	err := CreateDir(dir)
	if err != nil {
		t.Error(err.Error())
	} else {
		empty, err := IsEmptyDir(dir)
		if err != nil {
			t.Error(err.Error())
		}

		if empty == false {
			t.Error("结果错误，目录为空")
		}
	}
	os.Remove(dir)
}

func TestCreateDir(t *testing.T) {
	err := CreateDir(dir)
	if err != nil {
		t.Error(err.Error())
	} else {
		exists, err := IsExistsDir(dir)
		if err != nil {
			t.Error(err.Error())
		} else {
			if !exists {
				t.Error("结果错误，创建目录失败")
			} else {
				os.Remove(dir)
			}
		}
	}
}

func TestNotExistsDir(t *testing.T) {
	dir := "./test_dir1"
	exists, err := IsExistsDir(dir)
	if err != nil {
		t.Error(err.Error())
	}

	if exists == true {
		t.Error("结果错误，目录不存在")
	}
}

func TestCreateFile(t *testing.T) {
	file := "./test.txt"
	_, err := CreateFile(file)
	if err != nil {
		t.Error(err.Error())
	} else {
		exists, err := IsExistsFile(file)
		if err != nil {
			t.Error(err.Error())
		} else {
			if exists == false {
				t.Error("结果错误，文件创建失败")
			} else {
				os.Remove(file)
			}
		}
	}
}

func TestGetCurrentDir(t *testing.T) {
	currentDir := GetCurrentDir()
	assert.NotEmpty(t, currentDir)
}

func TestGetParDir(t *testing.T) {
	assert.NotEmpty(t, GetParDir())
}

func TestBackUpFile(t *testing.T) {
	log.NewLogger()
	err := EsimBackUpFile(file)
	assert.NoError(t, err)
}

func TestEsimRecoverFile(t *testing.T) {
	log.NewLogger()
	err := EsimBackUpFile(file)
	assert.NoError(t, err)
	err = EsimRecoverFile(file)
	assert.NoError(t, err)
}

func TestReadDir(t *testing.T) {
	paths, err := ReadDir("./")
	assert.Nil(t, err)
	assert.Equal(t, paths[0], "./test")

	paths, err = ReadDir(".")
	assert.Nil(t, err)
	assert.Equal(t, paths[0], "./test")

	paths, err = ReadDir("")
	assert.NotNil(t, err)
	assert.Nil(t, paths)
}
