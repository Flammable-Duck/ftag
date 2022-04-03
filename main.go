package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/alexflint/go-arg"
)

type AddCmd struct {
    Path string `arg:"positional" help:"path to the file to be added to ftag"`
    Tags []string `arg:"-t, required" help:"tags to tag the file with"`
}

type QueryCmd struct {
    Tags []string `arg:"positional" help:"tags to search for"`
}

var args struct {
    Add *AddCmd `arg:"subcommand:add" help:"add a file to ftag"`
    Query *QueryCmd `arg:"subcommand:query" help:"search for a tag"`
    File string  `arg:"-f" default:".ftag" help:"use a non-default ftag file"`
}

type File struct {
    Path string `json:"location"`
    Tags []string `json:"tags"`
}

func (f *File) addTag(tag string) {
    for i, oldTag := range f.Tags {
        if (oldTag == tag) {
            f.Tags = append(f.Tags[:i], f.Tags[i + 1:]...)
        }
    }
    f.Tags = append(f.Tags, tag)
}

func (f *File) hasTag(tag string) bool {
    for _, t := range f.Tags {
        if (t == tag) {
            return true
        }
    }
    return false
}

func save(files []File, dataFilePath string) {

    res,err := json.Marshal(files)
    if err != nil {
        fmt.Println(err)
    }

    ioutil.WriteFile(dataFilePath, res, 0666)
}

func load(dataFilePath string) ([]File, error) {
    var files []File
    f, err := ioutil.ReadFile(dataFilePath)
    if (err != nil) {
        return nil, err
    }

    err = json.Unmarshal(f, &files)

    if (err != nil) {
        if (err == fmt.Errorf("unexpected end of JSON input")) {
            return files, nil
        } else {
            return nil, err
        }
    }

    return files, err
}

func addFile(files []File, newFile File) []File {
    for i, file := range files {
        if (file.Path == newFile.Path) {
            files = append(files[:i], files[i + 1:]...)
            for _, tag := range file.Tags {
                newFile.addTag(tag)
            }
        }
    }
    files = append(files, newFile)
    return files
}

func tagQuery(files []File, tag string) []File {
    var result []File

    for _, f := range files {
        if (f.hasTag(tag)) {
            result = append(result, f)
        }
    }
    return result
}

func main() {

    p := arg.MustParse(&args)
    if p.Subcommand() == nil {
        p.Fail("missing subcommand")
    }

    // if (args.File == "") {
    //     args.File = ".ftag"
    // }

    files, err := load(args.File)
    if (err != nil) {
        if (err == fmt.Errorf("open .ftag: no such file or directory")) {
            fmt.Fprintf(os.Stderr, "ftag not initialized in this directory")
        } else {
            fmt.Fprintf(os.Stderr, "Error: %s", err)
        }
    }

    switch {
    case args.Add != nil:
        files = addFile(files, File{args.Add.Path, args.Add.Tags})
        save(files, args.File)
    case args.Query != nil:
        for _, f := range tagQuery(files, args.Query.Tags[0]) {
            fmt.Printf("%s\t%s\n", f.Path, f.Tags)
        }
    }
}
