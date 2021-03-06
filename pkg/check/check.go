package check

import (
	"fmt"
	"github.com/dkoshkin/invoices-validator/pkg/stringsx"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/thoas/go-funk"

	"github.com/dkoshkin/invoices-validator/pkg/validator"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"

	log "github.com/sirupsen/logrus"
)

const (
	// Do not allow ',' as that likely means its Last, First name
	folderNameRegex    = "^[^,]*$"
	folderNameExpected = "First Last, ie John Doe"

	fileNameRegex    = `^(0[1-9]|1[0-2])(0[1-9]|1\d|2\d|3[01])(1|2)\d{1}-\d{2}\.docx$`
	fileNameExpected = "Something like \"013119-01.docx\""
)

type Checker interface {
	Check(validator validator.Validator)
}

type FolderCheck struct {
	Folder          *files.FolderMetadata
	FoldersToIgnore []string
}

func (c *FolderCheck) Check(validator *validator.Validator) {
	name := c.Folder.Name

	// skip processing certain folders
	folders := stringsx.Split(filepath.Dir(c.Folder.PathLower), string(filepath.Separator))
	if len(funk.IntersectString(c.FoldersToIgnore, folders)) > 0 || funk.ContainsString(c.FoldersToIgnore, strings.ToLower(name)) {
		log.Debug("Ignoring Folder %q", name)
		return
	}

	log.Debug("Found Folder: %q", name)

	check := nameValidator{
		name:           name,
		regex:          folderNameRegex,
		expected:       folderNameExpected,
		additionalInfo: fmt.Sprintf("Folder: %q", c.Folder.PathDisplay),
	}
	validator.Validate(check)
}

type FileCheck struct {
	File            *files.FileMetadata
	FoldersToIgnore []string
	FilesToIgnore   []string
}

func (c *FileCheck) Check(validator *validator.Validator) {
	name := c.File.Name

	// skip processing certain files and files in skipped folders
	folders := stringsx.Split(filepath.Dir(c.File.PathLower), string(filepath.Separator))

	// if any parent Folder is in FOLDERS_TO_IGNORE OR File in FILES_TO_IGNORE, skip
	if len(funk.IntersectString(c.FoldersToIgnore, folders)) > 0 || funk.ContainsString(c.FilesToIgnore, name) {
		log.Debug("Ignoring File %q", c.File.PathDisplay)
		return
	}

	log.Debug("Found File: %q", c.File.PathDisplay)

	check := nameValidator{
		name:           name,
		regex:          fileNameRegex,
		expected:       fileNameExpected,
		additionalInfo: fmt.Sprintf("File: %q", c.File.PathDisplay),
	}
	validator.Validate(check)
}

type nameValidator struct {
	name           string
	regex          string
	expected       string
	additionalInfo string
}

func (nv nameValidator) Validate() (bool, []validator.ValidationError) {
	v := validator.NewValidator()

	match, _ := regexp.MatchString(nv.regex, nv.name)
	if !match {
		err := validator.ValidationError{
			Actual:         fmt.Sprintf("%q", nv.name),
			Expected:       nv.expected,
			AdditionalInfo: nv.additionalInfo,
		}
		v.AddError(err)
	}

	return v.Valid()
}
