package controller

import (
	"fmt"
	"github.com/dkoshkin/invoices-validator/pkg/check"
	"github.com/dkoshkin/invoices-validator/pkg/notifier"
	"github.com/dkoshkin/invoices-validator/pkg/stringsx"
	"github.com/dkoshkin/invoices-validator/pkg/validator"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"
	log "github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"time"

	"os"
	"strings"
)

const (
	dropboxTokenEnv = "DROPBOX_TOKEN"
	dropboxPathEnv  = "DROPBOX_PATH"

	foldersToIgnoreEnv = "FOLDERS_TO_IGNORE"
	filesToIgnoreEnv   = "FILES_TO_IGNORE"

	notifierSubjectBase = "Failed Invoice Validations"
)

func Run() error {
	// validate env vars are set
	dropboxToken := os.Getenv(dropboxTokenEnv)
	dropboxPath := os.Getenv(dropboxPathEnv)

	for env, val := range map[string]string{
		dropboxTokenEnv: dropboxToken,
		dropboxPathEnv:  dropboxPath} {
		if val == "" {
			return fmt.Errorf("%s variable must be set", env)
		}
	}

	notifiers, err := notifier.ConfiguredNotifiers()
	if err != nil {
		return fmt.Errorf("could not configure notifiers: %v", err)
	}

	// print some passed in env vars
	log.Infof("Using path: %q", dropboxPath)
	foldersToIgnore := stringsx.Split(os.Getenv(foldersToIgnoreEnv), ":")
	foldersToIgnoreLower := make([]string, 0, len(foldersToIgnore))
	funk.ForEach(foldersToIgnore, func(x string) {
		foldersToIgnoreLower = append(foldersToIgnoreLower, strings.ToLower(x))
	})
	log.Infof("Ignoring folders: %+v", foldersToIgnore)
	filesToIgnore := stringsx.Split(os.Getenv(filesToIgnoreEnv), ":")
	log.Infof("Ignoring file: %+v", filesToIgnore)

	// setup Dropbox client
	config := dropbox.Config{
		Token: dropboxToken,
		//LogLevel: dropbox.LogDebug, // if needed, set the desired logging level. Default is off
	}
	dbf := files.New(config)

	v := validator.NewValidator()
	hasMore := true
	cursor := ""
	for hasMore {
		// first get the parent folder and a cursor,
		// then use the cursor to get remaining files and folders
		res := &files.ListFolderResult{}
		if cursor == "" {
			in := &files.ListFolderArg{
				Path:      dropboxPath,
				Recursive: true,
			}
			var err error
			res, err = dbf.ListFolder(in)
			if err != nil {
				log.Errorf("could not list folders: %v", err)
			}
			cursor = res.Cursor
		} else {
			in := &files.ListFolderContinueArg{
				Cursor: cursor,
			}
			var err error
			res, err = dbf.ListFolderContinue(in)
			if err != nil {
				log.Errorf("could not list folders: %v", err)
			}
			cursor = res.Cursor
		}
		log.Debugf("Found %d files/folders in the directory", len(res.Entries))
		hasMore = res.HasMore
		for _, metadata := range res.Entries {
			switch metadata.(type) {
			// Casting types to access the metadata
			case *files.FolderMetadata:
				folder, _ := metadata.(*files.FolderMetadata)
				c := check.FolderCheck{
					Folder:          folder,
					FoldersToIgnore: foldersToIgnoreLower,
				}
				c.Check(v)
			case *files.FileMetadata:
				file, _ := metadata.(*files.FileMetadata)
				c := check.FileCheck{
					File:            file,
					FoldersToIgnore: foldersToIgnoreLower,
					FilesToIgnore:   filesToIgnore,
				}
				c.Check(v)
			}
		}
	}

	valid, errs := v.Valid()
	if !valid {
		log.Infof("Found %d errors", len(errs))

		for _, n := range notifiers {
			content, err := n.FormatContent(errs)
			if err != nil {
				log.Errorf("could not format content: %v", err)
			} else {
				notifierSubject := fmt.Sprintf("%s - %s", notifierSubjectBase, time.Now().Format("01022006"))
				err = n.Send(notifierSubject, content)
				if err != nil {
					log.Errorf("could not send notification: %v", err)
				}
			}
		}
	}

	return nil
}
