package main

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"

	"./checker"
)

type Summary struct {
	Inode   uint64
	Ctime   time.Time
	UID     string
	GID     string
	Mode    os.FileMode
	Links   uint64
	Owner   string
	Group   string
	Size    int64
	Blocks  int
	ModTime time.Time
	Name    string
	Dir     bool
}

/*InfoColor section used to change output text color in order of type of file selected*/
const InfoColor = "\033[1;34m%s\033[0m"
const InfoExecColor = "\033[01;32m%s\033[0m"
const InfoLinkColor = "\033[01;36m%s\033[0m"
const InfoSetGidColor = "\033[01;43m%s\033[0m"
const InfoSetUidColor = "\033[01;41m%s\033[0m"

var Info []Summary
var s Summary
var filesArr []string
var foldersArr []string

func main() {
	input := os.Args[1:]
	checker.Parser(input)
	for _, val := range checker.Input.Errors {
		fmt.Println(val)
	}
	if len(checker.Input.Path) == 0 {
		return
	}
	filesArr, foldersArr = DirSeparator()
	switch {
	case len(filesArr) != 0:
		for index, value := range filesArr {
			LSF(value, index)
			Info = []Summary{}
		}
		fallthrough
	case len(foldersArr) != 0:
		for index, value := range foldersArr {
			LS(value, index)
			Info = []Summary{}
		}
	}

}
func DirSeparator() (filesArr, foldersArr []string) {
	var file os.FileInfo
	for _, value := range checker.Input.Path {
		switch {
		case checker.Option['l']:
			file, _ = os.Lstat(value)
		default:
			file, _ = os.Stat(value)
		}
		mode := file.IsDir()
		switch {
		case !mode:
			filesArr = append(filesArr, value)
		case mode:
			foldersArr = append(foldersArr, value)
		}
	}
	return filesArr, foldersArr
}

func LSF(path string, n int) {
	fiInfo := []os.FileInfo{}
	switch {
	case path[0] != '/':
		tempPath, _ := os.Getwd()
		fiInfo = Harvest(tempPath, path)
		FillStruct(fiInfo, "")
	case path[0] == '/':
		index := strings.LastIndex(path, "/")
		tempPath := path[:index+1]
		fiInfo = Harvest(tempPath, path[index+1:])
		FillStruct(fiInfo, "")
		Info[0].Name = path
	}
	switch {
	case checker.Option['l']:
		switch {
		case n == len(filesArr)-1 && len(foldersArr) != 0:
			PrintStruct(Info, path)
			fmt.Println()
		default:
			PrintStruct(Info, path)
		}
	case len(filesArr) == 1 || n == len(filesArr)-1:
		if Info[0].Mode&os.ModeSymlink != 0 {
			fmt.Printf(InfoLinkColor, path+"\n")
		} else {
			fmt.Print(path + "\n")
		}
		switch {
		case len(foldersArr) > 0:
			fmt.Println()
		}
	case n < len(filesArr)-1:
		if Info[0].Mode&os.ModeSymlink != 0 {
			fmt.Printf(InfoLinkColor, path+"  ")
		} else {
			fmt.Print(path + "  ")
		}
	}
}
func LS(path string, n int) {
	fileInfo := Harvest(path, "")
	switch {
	case checker.Option['a']:
		LookUp(path)
	}
	newLine := ""
	if !LastRecord(n) {
		newLine = "\n"
	}
	switch {
	default:
		switch {
		case len(filesArr) > 0:
			fmt.Print(path + ":\n")
		case len(foldersArr) > 1 || len(checker.Input.Errors) != 0:
			fmt.Print(path + ":\n")
		}
		PrintStruct(Info, path)
		fmt.Print(newLine)
	case checker.Option['l'] && !checker.Option['R']:
		switch {
		case len(filesArr) > 0:
			fmt.Print(path + ":\n")
		case len(foldersArr) > 1:
			fmt.Print(path + ":\n")
		}
		fmt.Println("total", s.Total())
		PrintStruct(Info, path)
		fmt.Print(newLine)
	case checker.Option['R']:
		if LastRecord(n) && len(foldersArr) > 1 {
			fmt.Println()
		}
		fmt.Print(path + ":\n")
		switch {
		case checker.Option['l']:
			fmt.Println("total", s.Total())
		}
		PrintStruct(Info, path)
		Info = []Summary{}
		RangePath(path, fileInfo, "TraverseOff")
	}
}

/*ASAP - optimizing || refactoring*/
func LookUp(path string) {
	directRaw := []string{}
	direct := []string{}
	getPath := ""
	mode := "hidden"
	if path == "." || path[0] == '.' {
		getPath, _ = os.Getwd()
		getPath = getPath + path[1:]
		directRaw = strings.Split(strings.ReplaceAll(getPath, "/", " "), " ")
	} else if path[0] != '/' {
		getPath, _ = os.Getwd()
		getPath = getPath + "/" + path
		directRaw = strings.Split(strings.ReplaceAll(getPath, "/", " "), " ")
	} else {
		directRaw = strings.Split(strings.ReplaceAll(path, "/", " "), " ")
	}
	levelUP := "/"
	level2UP := "/"
	nameUP := ""
	name2UP := ""
	for _, val := range directRaw {
		if val != "" {
			direct = append(direct, val)
		}
	}
	for index, value := range direct {
		if index < len(direct)-1 {
			levelUP += value + "/"
		}
		if index < len(direct)-2 {
			level2UP += value + "/"
		}
	}
	if len(direct) == 1 {
		nameUP = direct[len(direct)-1]
		name2UP = nameUP
	}
	if len(direct) >= 2 {
		nameUP = direct[len(direct)-1]
		name2UP = direct[len(direct)-2]
	}
	fileInfo := []os.FileInfo{}
	dirInfo := Harvest(levelUP, nameUP)
	grdirInfo := Harvest(level2UP, name2UP)
	fileInfo = append(fileInfo, dirInfo...)
	fileInfo = append(fileInfo, grdirInfo...)
	FillStruct(fileInfo, mode)
}

func Harvest(path, dirname string) []os.FileInfo {
	dir, _ := os.Open(path)
	fileInfo, _ := dir.Readdir(0)
	dir.Close()
	mode := "normal"
	switch {
	case dirname != "":
		dirInfo := []os.FileInfo{}
		for _, file := range fileInfo {
			if file.Name() == dirname {
				dirInfo = append(dirInfo, file)
				return dirInfo
			}
		}
	}
	FillStruct(fileInfo, mode)
	return fileInfo
}

func LastRecord(n int) bool {
	if n < len(foldersArr)-1 {
		return false
	}
	return true
}

func PathBuilder(fileInfo []os.FileInfo) []os.FileInfo {
	var tempResult []os.FileInfo
	for _, file := range fileInfo {
		if file.Name()[:1] == "." && !checker.Option['a'] {
			continue
		}
		if !file.IsDir() {
			continue
		}
		tempResult = append(tempResult, file)
	}
	return tempResult
}

func RangePath(path string, fileInfo []os.FileInfo, mode string) {
	tempPath := PathBuilder(fileInfo)
	QuickSortNames(&tempPath, 0, len(tempPath)-1)
	PathWay := &tempPath
	for index, val := range *PathWay {
		if index <= len(*PathWay)-1 {
			fmt.Println()
		}
		tempPath := ""
		switch {
		case mode == "TraverseOn":
			tempPath = path + "/" + val.Name()
		case mode == "TraverseOff":
			switch {
			case path[len(path)-1:] != "/":
				tempPath = path + "/" + val.Name()
			default:
				tempPath = path + val.Name()
			}
		}
		Traverse(tempPath)
	}
}

func Traverse(path string) {
	fileInfo := Harvest(path, "")
	switch {
	case checker.Option['a']:
		LookUp(path)
	}
	fmt.Print(path + ":\n")
	switch {
	case checker.Option['l']:
		fmt.Println("total", s.Total())
	}
	PrintStruct(Info, path)
	Info = []Summary{}
	RangePath(path, fileInfo, "TraverseOn")
}

func FillStruct(fileInfo []os.FileInfo, mode string) {
	for index, file := range fileInfo {
		tempName := file.Name()
		if file.Name()[:1] == "." && !checker.Option['a'] {
			continue
		}
		switch {
		case mode == "hidden":
			switch {
			case index == 0:
				tempName = "."
			case index == 1:
				tempName = ".."
			}
		}
		links, uid, gid, blocks, ctime, inode := GetLinks(file)
		inf := Summary{
			Inode:   inode,
			Ctime:   ctime,
			UID:     uid,
			GID:     gid,
			Mode:    file.Mode(),
			Links:   links,
			Size:    file.Size(),
			Blocks:  blocks,
			ModTime: file.ModTime(),
			Name:    tempName,
			Dir:     file.IsDir(),
		}
		Info = append(Info, inf)
	}
}

func Strip(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		b := s[i]
		if ('a' <= b && b <= 'z') ||
			('A' <= b && b <= 'Z') ||
			('0' <= b && b <= '9') ||
			b == ' ' {
			result.WriteByte(b)
		}
	}
	return result.String()
}

func Partition(Info *[]Summary, lo, hi int) int {
	var a, b string
	temp := *Info
	for j := lo; j < hi; j++ {
		a = strings.ToUpper(Strip(temp[j].Name))
		b = strings.ToUpper(Strip(temp[hi].Name))
		switch {
		case checker.Option['t']:
			ctime1 := temp[j].Ctime
			ctime2 := temp[hi].Ctime
			unixTime1 := temp[j].ModTime.UnixNano()
			unixTime2 := temp[hi].ModTime.UnixNano()
			switch {
			default:
				if unixTime1 > unixTime2 {
					temp[j], temp[lo] = temp[lo], temp[j]
					lo++
				} else if unixTime1 == unixTime2 {
					if ctime1.Before(ctime2) {
						temp[j], temp[lo] = temp[lo], temp[j]
						lo++
					} else if ctime1 == ctime2 {
						if a < b {
							temp[j], temp[lo] = temp[lo], temp[j]
							lo++
						}
					}
				}
			case checker.Option['r']:
				if unixTime1 < unixTime2 {
					temp[j], temp[lo] = temp[lo], temp[j]
					lo++
				} else if unixTime1 == unixTime2 {
					if ctime2.Before(ctime1) {
						temp[j], temp[lo] = temp[lo], temp[j]
						lo++
					} else if ctime1 == ctime2 {
						if a > b {
							temp[j], temp[lo] = temp[lo], temp[j]
							lo++
						}
					}
				}
			}
		case checker.Option['r']:
			if a > b {
				temp[j], temp[lo] = temp[lo], temp[j]
				lo++
			}
		default:
			if a <= b {
				temp[j], temp[lo] = temp[lo], temp[j]
				lo++
			}
		}
	}
	temp[lo], temp[hi] = temp[hi], temp[lo]
	return lo
}

func QuickSort(Info *[]Summary, lo, hi int) {
	if lo > hi {
		return
	}
	a := *Info
	pivot := Partition(&a, lo, hi)
	QuickSort(&a, lo, pivot-1)
	QuickSort(&a, pivot+1, hi)

}

func PartitionNames(arr *[]os.FileInfo, lo, hi int) (*[]os.FileInfo, int) {
	var a, b string
	tempArr := *arr
	for j := lo; j < hi; j++ {
		a = strings.ToUpper(Strip(tempArr[j].Name()))
		b = strings.ToUpper(Strip(tempArr[hi].Name()))
		switch {
		case checker.Option['t']:
			_, _, _, _, ctime1, _ := GetLinks(tempArr[j])
			_, _, _, _, ctime2, _ := GetLinks(tempArr[hi])
			unixTime1 := tempArr[j].ModTime().UnixNano()
			unixTime2 := tempArr[hi].ModTime().UnixNano()
			switch {
			default:
				if unixTime1 > unixTime2 {
					tempArr[j], tempArr[lo] = tempArr[lo], tempArr[j]
					lo++
				} else if unixTime1 == unixTime2 {
					if ctime1.Before(ctime2) {
						tempArr[j], tempArr[lo] = tempArr[lo], tempArr[j]
						lo++
					} else if ctime1 == ctime2 {
						if a < b {
							tempArr[j], tempArr[lo] = tempArr[lo], tempArr[j]
							lo++
						}
					}
				}
			case checker.Option['r']:
				if unixTime1 < unixTime2 {
					tempArr[j], tempArr[lo] = tempArr[lo], tempArr[j]
					lo++
				} else if unixTime1 == unixTime2 {
					if ctime2.Before(ctime1) {
						tempArr[j], tempArr[lo] = tempArr[lo], tempArr[j]
						lo++
					} else if ctime1 == ctime2 {
						if a > b {
							tempArr[j], tempArr[lo] = tempArr[lo], tempArr[j]
							lo++
						}
					}
				}

			}
		case checker.Option['r']:
			if a > b {
				tempArr[j], tempArr[lo] = tempArr[lo], tempArr[j]
				lo++
			}
		case a <= b:
			tempArr[j], tempArr[lo] = tempArr[lo], tempArr[j]
			lo++
		}
	}
	tempArr[lo], tempArr[hi] = tempArr[hi], tempArr[lo]
	return &tempArr, lo
}
func QuickSortNames(arr *[]os.FileInfo, lo, hi int) {
	if lo > hi {
		return
	}
	a := *arr
	_, pivot := PartitionNames(&a, lo, hi)
	QuickSortNames(&a, lo, pivot-1)
	QuickSortNames(&a, pivot+1, hi)

}
func getTime(file time.Time) string {
	year, month, day := time.Now().Date()
	timeNow := time.Date(year, month-6, day, 0, 0, 0, 0, time.UTC)
	fileYear, fileMonth, fileDay := file.Date()
	timeFile := time.Date(fileYear, fileMonth, fileDay, 0, 0, 0, 0, time.UTC)
	arr := strings.Split(file.Format("Jan 2 15:04"), " ")
	arrYear := strings.Split(file.Format("Jan 2 2006"), " ")
	if timeNow.Before(timeFile) {
		if len(arr[1]) == 1 {
			return " " + arr[0] + "  " + arr[1] + " " + arr[2] + " "
		} else {
			return " " + arr[0] + " " + arr[1] + " " + arr[2] + " "
		}
	} else {
		if len(arrYear[1]) == 2 {
			return " " + arrYear[0] + " " + arrYear[1] + "  " + arrYear[2] + " "
		} else {
			return " " + arrYear[0] + "  " + arrYear[1] + "  " + arrYear[2] + " "
		}
	}
}

func PrintStruct(arr []Summary, path string) {
	// arr := *arrPoint
	if len(arr) == 0 {
		return
	}
	largest, largest1 := largestSize(arr)
	whiteSpaces := ""
	whiteSpaces1 := ""
	QuickSort(&Info, 0, len(Info)-1)
	for _, val := range arr {
		switch {
		case checker.Option['l']:
			current := len(strconv.Itoa(int(val.Size)))
			current1 := len(val.GID)
			for i := 0; i < int(largest)-current; i++ {
				whiteSpaces += " "
			}
			for i := 0; i < int(largest1)-current1; i++ {
				whiteSpaces1 += " "
			}
			groupname := val.GID
			username := val.UID
			fmt.Print(val.Mode, val.Links, " ", username, " ", groupname+whiteSpaces1, " "+whiteSpaces, val.Size, getTime(val.ModTime), " ")
			whiteSpaces = ""
			whiteSpaces1 = ""
		}

		if !val.Dir && val.Mode&os.ModePerm >= 444 && val.Mode&os.ModeSymlink == 0 {
			if val.Mode&os.ModeSetgid != 0 {
				fmt.Printf(InfoSetGidColor, val.Name)
			} else if val.Mode&os.ModeSetuid != 0 {
				fmt.Printf(InfoSetUidColor, val.Name)
			} else {
				fmt.Printf(InfoExecColor, val.Name+" ")
			}
		} else if !val.Dir && val.Mode&os.ModeSymlink != 0 {
			symLink := ""
			var err error
			switch {
			case path == val.Name:
				symLink, err = os.Readlink(val.Name)
			default:
				symLink, err = os.Readlink(path + "/" + val.Name)
			}
			if err != nil {
				fmt.Println(err)
			}
			fmt.Printf(InfoLinkColor, val.Name+" ")
			if checker.Option['l'] {
				fmt.Print("-> ")
				if val.Mode&os.ModePerm >= 444 {
					fmt.Printf(InfoExecColor, symLink)
				} else {
					fmt.Print(symLink)
				}
			}
		} else if val.Dir {
			fmt.Printf(InfoColor, val.Name+" ")
		} else {
			fmt.Printf(val.Name + " ")
		}
		switch {
		case checker.Option['l']:
			fmt.Println()
		}
	}
	switch {
	case !checker.Option['l']:
		fmt.Println()
	}
}

func largestSize(arr []Summary) (int, int) {
	var max, maxgrlen int64
	for _, val := range arr {
		if val.Size > max {
			max = val.Size
		}
		if len(val.GID) > int(maxgrlen) {
			maxgrlen = int64(len(val.GID))
		}
	}
	return len(strconv.Itoa(int(max))), int(maxgrlen)
}

func (s *Summary) Total() int {
	total := 0
	for i := 0; i < len(Info); i++ {
		total += Info[i].Blocks
	}
	return (total + 1) / 2
}

func GetLinks(file os.FileInfo) (uint64, string, string, int, time.Time, uint64) {
	var uid, gid, blocks int
	var inode uint64
	var ctime time.Time
	nlink := uint64(0)
	if sys := file.Sys(); sys != nil {
		if stat, ok := sys.(*syscall.Stat_t); ok {
			ctime = time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
			inode = stat.Ino
			nlink = uint64(stat.Nlink)
			uid = int(stat.Uid)
			gid = int(stat.Gid)
			blocks = int(stat.Blocks)
		}
	}
	uidString, _ := user.LookupGroupId(strconv.Itoa(uid))
	gidString, _ := user.LookupGroupId(strconv.Itoa(gid))
	return nlink, uidString.Name, gidString.Name, blocks, ctime, inode
}
