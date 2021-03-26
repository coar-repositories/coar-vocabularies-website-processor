package main

import (
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
	"io"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

func resetVolatileFolder(folderPath string) error {
	var err error
	err = os.RemoveAll(folderPath)
	if err != nil {
		return err
	}
	err = os.MkdirAll(folderPath, os.ModePerm)
	return err
}

func getConceptNameFromUri(namespaceUri, conceptUri string) string {
	return strings.TrimPrefix(conceptUri, namespaceUri)
}

func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

func copyFile(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}
	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}
	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()
	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func regexReplaceInFile(filePath, regexString, replaceString string) error {
	var err error
	content, err := ioutil.ReadFile(filePath)
	regex := regexp.MustCompile(regexString)
	content = regex.ReplaceAll(content, []byte(replaceString))
	err = ioutil.WriteFile(filePath, content, os.ModePerm)
	return err
}

func languageTagFromLiteral(serialisedLiteral string) string {
	languageCode := regexp.MustCompile(`@\w\w$`).FindString(serialisedLiteral)
	if languageCode == "" {
		zapLogger.Warn("serialised literal does not have language code", zap.String("literal", serialisedLiteral))
		return languageCode
	} else {
		return languageCode[1:len(languageCode)]
	}
}

func languageNameFromTag(languageTag string) string {
	t := language.MustParse(languageTag)
	return display.Self.Name(t)
}
