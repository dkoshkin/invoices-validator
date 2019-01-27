package main

import (
	"flag"
	"github.com/dkoshkin/invoices-validator/pkg/check"
	"github.com/dkoshkin/invoices-validator/pkg/notifier"
	"github.com/dkoshkin/invoices-validator/pkg/stringsx"
	"github.com/dkoshkin/invoices-validator/pkg/validator"
	"os"
	"strings"

	"github.com/thoas/go-funk"

	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/dropbox/files"

	log "k8s.io/klog"
)

const (
	dropboxTokenEnv = "DROPBOX_TOKEN"
	dropboxPathEnv  = "DROPBOX_PATH"

	foldersToIgnoreEnv = "FOLDERS_TO_IGNORE"
	filesToIgnoreEnv   = "FILES_TO_IGNORE"

	notifierSubject = "Failed Invoice Validations"
)

// Set via linker flag
var (
	version   string
	buildDate string
)

func main() {
	// setup logging
	log.InitFlags(nil)
	flag.Parse()

	log.Infof("Version: %q", version)
	log.Infof("Build Date: %q", buildDate)

	// validate env vars are set
	dropboxToken := os.Getenv(dropboxTokenEnv)
	dropboxPath := os.Getenv(dropboxPathEnv)

	for env, val := range map[string]string{
		dropboxTokenEnv: dropboxToken,
		dropboxPathEnv:  dropboxPath} {
		if val == "" {
			log.Fatalf("%s variable must be set", env)
		}
	}

	notifiers, err := notifier.ConfiguredNotifiers()
	if err != nil {
		log.Fatalf("could not configure notifiers: %v", err)
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
		log.V(2).Infof("Found %d files/folders in the directory", len(res.Entries))
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
		log.Infof("Found %d errors, will send notification(s)", len(errs))

		for _, n := range notifiers {
			content, err := n.FormatContent(errs)
			if err != nil {
				log.Errorf("could not format content: %v", err)
			} else {
				err = n.Send(notifierSubject, content)
				if err != nil {
					log.Errorf("could not send notification: %v", err)
				}
			}
		}
	}
}
