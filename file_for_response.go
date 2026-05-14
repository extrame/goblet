package goblet

import (
	"io"
	"os"
)

type fileWithName struct {
	origin io.ReadSeeker
	size   int64
	name   string
}

// ReadSeekerStat 组合io.ReadSeeker和Stat接口
type ReadSeekerStat interface {
	io.ReadSeeker
	Stat() (os.FileInfo, error)
}

// getFileSize 获取文件大小
// 如果origin实现了Stat方法，则使用Stat
// 否则使用Seek方式获取
func getFileSize(file io.ReadSeeker) int64 {
	// 优先尝试转换为ReadSeekerStat（如果实现了Stat方法）
	if statFile, ok := file.(ReadSeekerStat); ok {
		if info, err := statFile.Stat(); err == nil {
			return info.Size()
		}
	}

	// 使用Seek方式获取文件大小
	// 保存当前位置
	currentPos, _ := file.Seek(0, io.SeekCurrent)
	// 移动到末尾获取大小
	size, _ := file.Seek(0, io.SeekEnd)
	// 恢复到原始位置
	file.Seek(currentPos, io.SeekStart)

	return size
}

// FileWithName make file can by download for another name
func FileWithName(file io.ReadSeeker, name string) *fileWithName {
	// 计算文件大小
	size := getFileSize(file)

	return &fileWithName{
		origin: file.(ReadSeekerStat), // 安全的类型断言，因为getFileSize已经验证过
		size:   size,
		name:   name,
	}
}

func (f *fileWithName) Read(p []byte) (n int, err error) {
	return f.origin.Read(p)
}

func (f *fileWithName) Seek(offset int64, whence int) (int64, error) {
	return f.origin.Seek(offset, whence)
}

func (f *fileWithName) GetName() string {
	return f.name
}

func (f *fileWithName) GetSize() int64 {
	return f.size
}
