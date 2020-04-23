package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"time"

	"github.com/Cloud-Foundations/Dominator/lib/log"
	"github.com/Cloud-Foundations/golib/pkg/auth/userinfo/gitdb"
)

var divider = []byte("======================================================\n")

func showUserGroupsSubcommand(args []string, logger log.DebugLogger) error {
	if err := showUserGroups(os.Stdout, args[0], args[1], logger); err != nil {
		return fmt.Errorf("Error showing groups for user: %s: %s", args[1], err)
	}
	return nil
}

func showUserGroups(writer io.Writer, source, username string,
	logger log.DebugLogger) error {
	var tmpdir string
	fi, err := os.Stat(source)
	if err == nil && fi.IsDir() {
		tmpdir = source
		source = ""
	} else {
		tmpdir, err = ioutil.TempDir("", "userinfo")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpdir)
	}
	memoryLogger := newMemoryLogger()
	if !*ignoreErrors {
		logger = memoryLogger
	}
	db, err := gitdb.New(source, "", tmpdir, time.Hour, logger)
	if err != nil {
		return err
	}
	if memoryLogger.buffer.Len() > 0 {
		if _, err := memoryLogger.buffer.WriteTo(os.Stderr); err != nil {
			return err
		}
		return errors.New("database not clean")
	}
	if username != "" {
		if groups, err := db.GetUserGroups(username); err != nil {
			return err
		} else {
			showLine(writer, username, groups)
			return nil
		}
	}
	usernames, err := db.GetUsersInGroups()
	if err != nil {
		return err
	}
	sort.Strings(usernames)
	for _, username := range usernames {
		if groups, err := db.GetUserGroups(username); err != nil {
			return err
		} else {
			showLine(writer, username, groups)
		}
	}
	writer.Write(divider)
	groups, err := db.GetGroups()
	if err != nil {
		return err
	}
	sort.Strings(groups)
	for _, group := range groups {
		if users, err := db.GetUsersInGroup(group); err != nil {
			return err
		} else {
			showLine(writer, group, users)
		}
	}
	return nil
}

func showLine(writer io.Writer, key string, values []string) {
	sort.Strings(values)
	fmt.Fprint(writer, key+":")
	for _, value := range values {
		fmt.Fprint(writer, " ", value)
	}
	fmt.Fprintln(writer)
}
