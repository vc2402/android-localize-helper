package engine

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const (
	defLocale   = "def"
	stringsFile = "strings.xml"
	valuesDir   = "values"
	nameColumn  = "id"
	xmlIndent   = "  "
)

type xStrings struct {
	XMLName xml.Name  `xml:"resources"`
	Strings []xString `xml:"string"`
}

type xString struct {
	Name         string `xml:"name,attr"`
	Value        string `xml:",chardata"`
	Translatable string `xml:"translatable,attr,omitempty"`
}

//String contains all the strings of project
type String struct {
	Name         string
	Values       map[string]string
	Translatable bool
}

//Localizer contains localization engine data
type Localizer struct {
	ResourcesDir string
	Locales      []string
	strings      map[string]*String
	err          error
}

//New creates new localization engine
func New(projectDir string, locales ...string) *Localizer {
	l := &Localizer{Locales: []string{defLocale}}
	resPath := filepath.Join(projectDir, "app/src/main/res")
	l.err = checkPathIsResourcesDir(resPath)
	if l.err != nil {
		resPath = projectDir
		if checkPathIsResourcesDir(resPath) != nil {
			return l
		}
	}
	l.ResourcesDir = resPath
	if len(locales) > 0 {
		l.addLocales(locales)
	} else {
		l.guessLocales()
	}
	return l
}

//AddLocale adds locale to localizer
func (l *Localizer) AddLocale(loc string) *Localizer {
	l.addLocale(loc)
	return l
}

//Load tries to parse resource files and store strings in engine structure
func (l *Localizer) Load() *Localizer {
	if l.err != nil {
		return l
	}
	l.strings = map[string]*String{}
	for _, loc := range l.Locales {
		fileName := l.getFileNameForLocale(loc, false)
		rf, err := l.readResources(fileName)
		if err != nil {
			l.err = err
			return l
		}
		for _, r := range rf.Strings {
			s, ok := l.strings[r.Name]
			if !ok {
				s = &String{Name: r.Name, Values: map[string]string{}, Translatable: true}
				l.strings[r.Name] = s
			}
			s.Values[loc] = r.Value
			if r.Translatable == "false" {
				s.Translatable = false
			}
		}
	}
	return l
}

//Save saves values to all non-default locales resources
func (l *Localizer) Save() error {
	if l.err != nil {
		return l.err
	}
	for _, loc := range l.Locales {
		if loc != defLocale {
			res := &xStrings{Strings: []xString{}}
			for n, s := range l.strings {
				if s.Translatable {
					v, ok := s.Values[loc]
					if !ok {
						v = s.Values[defLocale]
					}
					str := xString{Name: n, Value: v}
					res.Strings = append(res.Strings, str)
				}
			}
			fileName := l.getFileNameForLocale(loc, true)
			err := l.writeResources(fileName, res)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//Export exports data to csv file
func (l *Localizer) Export(fileName string) error {
	if l.err != nil {
		return l.err
	}
	of, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer of.Close()
	return l.ExportW(of)
}

//ExportW writes data in csv format to given writer
func (l *Localizer) ExportW(w io.Writer) (err error) {
	if l.err != nil {
		return l.err
	}
	cw := csv.NewWriter(w)
	row := make([]string, len(l.Locales)+1)
	row[0] = nameColumn
	for i, l := range l.Locales {
		row[i+1] = l
	}
	err = cw.Write(row)
	if err != nil {
		return
	}
	for k, s := range l.strings {
		if s.Translatable {
			row[0] = k
			for i, l := range l.Locales {
				row[i+1], _ = s.Values[l]
			}
			err = cw.Write(row)
			if err != nil {
				return
			}
		}
	}
	cw.Flush()
	return nil
}

//Import imports data from csv file
func (l *Localizer) Import(fileName string) error {
	if l.err != nil {
		return l.err
	}
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	return l.ImportR(f)
}

//ImportR imports values in csv format from reader
func (l *Localizer) ImportR(r io.Reader) (err error) {
	if l.err != nil {
		return l.err
	}
	cr := csv.NewReader(r)
	cr.ReuseRecord = true
	row, err := cr.Read()
	if err != nil {
		return err
	}
	if row[0] != nameColumn {
		return fmt.Errorf("invalid csv format: first column name should be '%s', not '%s'", nameColumn, row[0])
	}
	if row[1] != defLocale {
		return fmt.Errorf("invalid csv format: second column name should be '%s', not '%s'", defLocale, row[1])
	}
	locales := make([]string, len(row)-2)
	for i := 2; i < len(row); i++ {
		l.addLocale(row[i])
		locales[i-2] = row[i]
	}

	for {
		row, err = cr.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		s, ok := l.strings[row[0]]
		if !ok {
			return fmt.Errorf("value with name '%s' from csv is not found in resources file", row[0])
		}
		for i, loc := range locales {
			s.Values[loc] = row[i+2]
		}
	}
	return
}

//Strings returns imported strings slice
func (l *Localizer) Strings() map[string]*String {
	return l.strings
}

//Err returns engine's error
func (l *Localizer) Err() error {
	return l.err
}

func (l *Localizer) addLocales(ls []string) {
	for _, loc := range ls {
		l.addLocale(loc)
	}
}

func (l *Localizer) addLocale(loc string) {
	for _, lc := range l.Locales {
		if lc == loc {
			return
		}
	}
	l.Locales = append(l.Locales, loc)
}

func (l *Localizer) getFileNameForLocale(loc string, checkDir bool) string {
	if loc == defLocale {
		return filepath.Join(l.ResourcesDir, valuesDir, stringsFile)
	}
	dir := filepath.Join(l.ResourcesDir, valuesDir+"-"+loc)
	if checkDir {
		_, e := os.Stat(dir)
		if e != nil {
			os.Mkdir(dir, os.ModePerm)
		}
	}
	return filepath.Join(dir, stringsFile)
}

func (l *Localizer) readResources(fileName string) (resources *xStrings, err error) {
	f, err := os.Open(fileName)
	if err != nil {
		return
	}
	defer f.Close()
	byteValue, _ := ioutil.ReadAll(f)
	resources = &xStrings{}
	err = xml.Unmarshal(byteValue, resources)
	return
}

func (l *Localizer) writeResources(fileName string, resources *xStrings) (err error) {
	_, e := os.Stat(fileName)
	if e == nil {
		os.Rename(fileName, fileName+".bak")
	}
	f, err := os.Create(fileName)
	if err != nil {
		return
	}
	defer f.Close()
	var bytes []byte
	bytes, err = xml.MarshalIndent(resources, "", xmlIndent)
	if err == nil {
		f.Write(bytes)
	}
	return
}

func (l *Localizer) guessLocales() {
	templ := valuesDir + "-"
	files, err := ioutil.ReadDir(l.ResourcesDir)
	if err == nil {
		for _, f := range files {
			if f.IsDir() && strings.Index(f.Name(), templ) == 0 {
				l.Locales = append(l.Locales, f.Name()[len(templ):])
			}
		}
	}
}

func checkPathIsResourcesDir(p string) error {
	rs, err := os.Stat(p)
	if err == nil {
		if !rs.IsDir() {
			err = fmt.Errorf("%s is not a dir", p)
		} else {
			strFile := filepath.Join(p, valuesDir, stringsFile)
			rs, err = os.Stat(strFile)
		}
	}
	return err
}
