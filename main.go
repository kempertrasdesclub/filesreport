package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
)

type Grouping struct {
	Size int64
	Files int
}

type ComplexFileInfo struct {
	Name        string
	Path        string
	Extension   string
	Mime        string
	Size        int64
	IsDir       bool
	Created     time.Time
	Mode        fs.FileMode
}

type Year int
type Month int
type Day int
type Hour int
type Minute int
type Mime string

type FileInfo map[Year]map[Month]map[Day]map[Hour]map[Minute]map[Mime]Grouping

func (e *FileInfo) Set(year, month, day, hour, minute int, size int64, mime string) {
	var localGrouping Grouping
	var found bool
	
	if *e == nil {
		*e = make(map[Year]map[Month]map[Day]map[Hour]map[Minute]map[Mime]Grouping)
	}
	
	if (*e)[Year(year)] == nil {
		(*e)[Year(year)] = make(map[Month]map[Day]map[Hour]map[Minute]map[Mime]Grouping)
	}
	
	if (*e)[Year(year)][Month(month)] == nil {
		(*e)[Year(year)][Month(month)] = make(map[Day]map[Hour]map[Minute]map[Mime]Grouping)
	}
	
	if (*e)[Year(year)][Month(month)][Day(day)] == nil {
		(*e)[Year(year)][Month(month)][Day(day)] = make(map[Hour]map[Minute]map[Mime]Grouping)
	}
	
	if (*e)[Year(year)][Month(month)][Day(day)][Hour(hour)] == nil {
		(*e)[Year(year)][Month(month)][Day(day)][Hour(hour)] = make(map[Minute]map[Mime]Grouping)
	}
	
	if (*e)[Year(year)][Month(month)][Day(day)][Hour(hour)][Minute(minute)] == nil {
		(*e)[Year(year)][Month(month)][Day(day)][Hour(hour)][Minute(minute)] = make(map[Mime]Grouping)
	}
	
	localGrouping, found = (*e)[Year(year)][Month(month)][Day(day)][Hour(hour)][Minute(minute)][Mime(mime)]
	if found == true {
		localGrouping.Files += 1
		localGrouping.Size  += size
	} else {
		localGrouping = Grouping{
			Size:  size,
			Files: 1,
		}
	}
	
	(*e)[Year(year)][Month(month)][Day(day)][Hour(hour)][Minute(minute)][Mime(mime)] = localGrouping
}

const KSplitString = "|"

func main() {
	var err error
	var pathInfo []ComplexFileInfo
	var groupingInfo FileInfo
	var pathScanList []string
	var toPrint = make(map[string]FileInfo)
	var pathScan = os.Getenv("PATH_SCAN")
	var output = os.Getenv("OUTPUT")
	
	if len(os.Args) > 1 && strings.ToLower(os.Args[1]) == "help" {
		fmt.Printf("\n")
		fmt.Printf("Este programa varre diretórios e imprime um relatório em json\n")
		fmt.Printf("Forma de uso:\n\n")
		fmt.Printf("   envvar PATH_SCAN=dir_1"+KSplitString+"dir_2"+KSplitString+"dir_3"+KSplitString+"dir_N\n")
		fmt.Printf("   envvar OUTPUT=print               - Imprime o relatório na saída padrão;\n")
		fmt.Printf("   envvar OUTPUT=/dir/file_name.json - Imprime o relatório em um arquivo json\n\n")
		fmt.Printf("Formato do json:\n\n")
		fmt.Printf("   {\"path (string)\":{\"year (int)\":{\"month (int - 1:january, ...)\":{\"day (int)\":{\"hour (int)\":{\"minute (int)\":{\"type (int)\":{\"Size\":int,\"Files\":int}}}}}}}}")
		fmt.Printf("\n")
		os.Exit(0)
	}
	
	if pathScan == "" {
		fmt.Printf("Por favor, defina a variável de ambiente PATH_SCAN com todos os diretórios separados por ponto e vírgula\n")
		os.Exit(0)
	}
	
	if output == "" {
		output = "print"
		
		fmt.Printf("Por favor, defina a variável de ambiente OUTPUT como 'print' ou caminho e nome do arquivo. Ex.: './relatorio.json'\n")
		fmt.Printf("OUTPUT definido como 'print'\n")
	}
	
	pathScanList = strings.Split(pathScan, KSplitString)
	
	for _, path := range pathScanList {
		path = strings.Trim(path, " ")
		
		pathInfo = []ComplexFileInfo{}
		groupingInfo = FileInfo{}
		
		err = ScanDir(path, &pathInfo)
		if err != nil {
			panic(err)
		}
		
		CountFile(&pathInfo, &groupingInfo)
		
		toPrint[path] = groupingInfo
	}
	
	var jsonOut []byte
	jsonOut, err = json.Marshal(&toPrint)
	if err != nil {
		panic(err)
	}
	
	if strings.ToLower(output) == "print" {
		fmt.Printf("%s", jsonOut)
	} else {
		var f *os.File
		f, err = os.OpenFile(output, os.O_CREATE | os.O_WRONLY, os.ModePerm)
		if err != nil {
			panic(err)
		}
		_, err = f.Write(jsonOut)
		if err != nil {
			panic(err)
		}
		
		err = f.Close()
		if err != nil {
			panic(err)
		}
	}
	
	
}

func CountFile(pathInfo *[]ComplexFileInfo, groupingInfo *FileInfo) {
	for _, fileInfo := range *pathInfo {
		year    := fileInfo.Created.Year()
		month   := fileInfo.Created.Month()
		day     := fileInfo.Created.Day()
		hour    := fileInfo.Created.Hour()
		minute  := fileInfo.Created.Minute()
		mime    := fileInfo.Mime
		size    := fileInfo.Size
		
		groupingInfo.Set(year, int(month), day, hour, minute, size, mime)
	}
	
	return
}

func ScanDir(path string, data *[]ComplexFileInfo) (err error) {
	log.SetPrefix("ScanDir(): ")
	defer log.SetPrefix("")
	
	var complexFileInfo ComplexFileInfo
	var info []fs.FileInfo
	var buf  []byte
	var kind types.Type
	
	path, err = filepath.Abs(path)
	if err != nil {
		log.Printf("filepath.Abs().error: %v", err)
		return
	}
	
	info, err = ioutil.ReadDir(path)
	if err != nil {
		log.Printf("ioutil.ReadDir().error: %v", err)
		return
	}
	
	for _, v := range info {
		stop := v.Name()
		_ = stop
		
		complexFileInfo = ComplexFileInfo{}
		
		if v.IsDir() == true {
			var loopData []ComplexFileInfo
			
			complexFileInfo.Path  = filepath.Join(path, v.Name())
			complexFileInfo.IsDir = true
			
			err = ScanDir(filepath.Join(path, v.Name()), &loopData)
			if err != nil {
				log.Printf("ScanDir().error: %v", err)
				return
			}
			
			*data = append(*data, loopData...)
			
		} else {
			buf, err = ioutil.ReadFile(filepath.Join(path, v.Name()))
			if err != nil {
				log.Printf("ioutil.ReadFile().error: %v", err)
				return
			}
			
			if len(buf) == 0 {
				log.Printf("empty buffer: %v", filepath.Join(path, v.Name()))
				continue
			}
			
			kind, err = filetype.Match(buf)
			if err != nil {
				log.Printf("filetype.Match().error: %v", err)
				return
			}
			
			complexFileInfo.Name = v.Name()
			complexFileInfo.Path = path
			complexFileInfo.Size = v.Size()
			complexFileInfo.Created = v.ModTime()
			complexFileInfo.Mode = v.Mode()
			
			if kind == filetype.Unknown {
				var extensionTmp = strings.Split(v.Name(), ".")
				complexFileInfo.Extension = strings.ToLower(extensionTmp[len(extensionTmp)-1])
				complexFileInfo.Mime = "unknown"
				
				*data = append(*data, complexFileInfo)
				continue
			}
			
			complexFileInfo.Mime      = kind.MIME.Type
			complexFileInfo.Extension = kind.Extension
			complexFileInfo.Name      = v.Name()
			complexFileInfo.Path      = path
			
			*data = append(*data, complexFileInfo)
		}
	}
	
	return
}