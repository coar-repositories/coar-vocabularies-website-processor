package main

import (
	"fmt"
	"github.com/goki/ki/ki"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Website struct {
	WebrootFolderPath       string
	ContentFolderPath       string
	StaticContentFolderPath string
}

func (website *Website) Initialise(webPageSourcesPath, webrootPath string) error {
	var err error
	website.WebrootFolderPath = webrootPath
	website.ContentFolderPath = filepath.Join(website.WebrootFolderPath, "content")
	website.StaticContentFolderPath = filepath.Join(website.WebrootFolderPath, "static")
	err = resetVolatileFolder(website.ContentFolderPath)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	err = resetVolatileFolder(website.StaticContentFolderPath)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	_, err = copyFile(filepath.Join(webPageSourcesPath, "_index.md"), filepath.Join(website.ContentFolderPath, "_index.md"))
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	_, err = copyFile(filepath.Join(webPageSourcesPath, "about.md"), filepath.Join(website.ContentFolderPath, "about.md"))
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	return err
}

func (website *Website) ProcessConceptScheme(conceptScheme *ConceptScheme) error {
	var err error
	for _, conceptSchemeVersion := range conceptScheme.Versions {
		if conceptSchemeVersion.Version == conceptScheme.GetLatestVersion().Version {
			conceptSchemeVersion.Latest = true
		}
		err = website.ProcessConceptSchemeVersion(&conceptSchemeVersion, false)
		if err != nil {
			zapLogger.Error(err.Error())
			return err
		}
		if conceptSchemeVersion.Version == conceptScheme.GetLatestVersion().Version {
			err = website.ProcessConceptSchemeVersion(&conceptSchemeVersion, true)
			if err != nil {
				zapLogger.Error(err.Error())
				return err
			}
		}
	}
	currentConceptSchemeVersion := conceptScheme.GetLatestVersion()
	for _, conceptSchemeVersion := range conceptScheme.Versions {
		for _, concept := range conceptSchemeVersion.Concepts {
			if currentConceptSchemeVersion.GetConceptById(concept.ID) == nil {
				//	this is an orphan concept
				concept.HugoLayout = "concept_orphan"
				conceptPage, conceptMarshalErr := concept.marshal()
				if conceptMarshalErr != nil {
					zapLogger.Error(conceptMarshalErr.Error())
					return conceptMarshalErr
				}
				conceptPageFolderPath := filepath.Join(website.ContentFolderPath, conceptScheme.ID, concept.ID)
				zapLogger.Debug("Writing orphan concept page to ", zap.String("path", conceptPageFolderPath))
				os.MkdirAll(conceptPageFolderPath, os.ModePerm)
				conceptFileWriteErr := ioutil.WriteFile(filepath.Join(conceptPageFolderPath, "index.md"), conceptPage, os.ModePerm)
				if conceptFileWriteErr != nil {
					zapLogger.Error(conceptFileWriteErr.Error())
					return conceptFileWriteErr
				}
			}
		}
	}
	return err
}

func (website *Website) ProcessConceptSchemeVersion(conceptSchemeVersion *ConceptSchemeVersion, asCurrentVersion bool) error {
	var err error
	conceptSchemeVersion.Current = asCurrentVersion
	err = os.MkdirAll(conceptSchemeVersion.CalculateFolderPath(website.ContentFolderPath, asCurrentVersion), os.ModePerm)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	conceptSchemePage, conceptSchemeMarshalErr := conceptSchemeVersion.Marshal()
	if conceptSchemeMarshalErr != nil {
		zapLogger.Error(conceptSchemeMarshalErr.Error())
		return conceptSchemeMarshalErr
	}
	fileWriteErr := ioutil.WriteFile(filepath.Join(conceptSchemeVersion.CalculateFolderPath(website.ContentFolderPath, asCurrentVersion), "_index.md"), conceptSchemePage, os.ModePerm)
	if fileWriteErr != nil {
		zapLogger.Error(fileWriteErr.Error())
		return fileWriteErr
	}
	err = website.GenerateConceptPages(conceptSchemeVersion)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	err = website.GenerateHtmlTree(conceptSchemeVersion, asCurrentVersion)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	err = website.GeneratePrintableSinglePage(conceptSchemeVersion, asCurrentVersion)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	err = website.GenerateDownloadFiles(conceptSchemeVersion, asCurrentVersion)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	return err
}

func (website *Website) GeneratePrintableSinglePage(conceptSchemeVersion *ConceptSchemeVersion, asCurrentVersion bool) error {
	var err error
	conceptSchemeVersion.HugoLayout = "printable"
	printableVocabPage, conceptSchemeMarshalErr := conceptSchemeVersion.Marshal()
	if conceptSchemeMarshalErr != nil {
		zapLogger.Error(conceptSchemeMarshalErr.Error())
		return conceptSchemeMarshalErr
	}
	conceptSchemeVersion.HugoLayout = ""
	err = ioutil.WriteFile(filepath.Join(conceptSchemeVersion.CalculateFolderPath(website.ContentFolderPath, asCurrentVersion), "printable.md"), printableVocabPage, os.ModePerm)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	return err
}

func (website *Website) GenerateConceptPages(conceptSchemeVersion *ConceptSchemeVersion) error {
	var err error
	for _, concept := range conceptSchemeVersion.Concepts {
		conceptPage, conceptMarshalErr := concept.marshal()
		if conceptMarshalErr != nil {
			zapLogger.Error(conceptMarshalErr.Error())
			return conceptMarshalErr
		}
		conceptPageFolderPath := filepath.Join(conceptSchemeVersion.CalculateFolderPath(website.ContentFolderPath, true), concept.ID)
		os.MkdirAll(conceptPageFolderPath, os.ModePerm)
		conceptFileWriteErr := ioutil.WriteFile(filepath.Join(conceptPageFolderPath, "index.md"), conceptPage, os.ModePerm)
		if conceptFileWriteErr != nil {
			zapLogger.Error(conceptFileWriteErr.Error())
			return conceptFileWriteErr
		}
	}
	return err
}

func (website *Website) GenerateDownloadFiles(conceptSchemeVersion *ConceptSchemeVersion, asCurrentVersion bool) error {
	destinationFolderPath := filepath.Join(website.StaticContentFolderPath, conceptSchemeVersion.ID, conceptSchemeVersion.Version)
	if asCurrentVersion {
		destinationFolderPath = filepath.Join(website.StaticContentFolderPath, conceptSchemeVersion.ID)
	}
	err := os.MkdirAll(destinationFolderPath, os.ModePerm)
	if err != nil {
		zapLogger.Debug(err.Error())
		return err
	}
	_, err = copyFile(filepath.Join(conceptSchemeVersion.SkosProcessedFolderPath, conceptSchemeVersion.ID+"_for_dspace.xml"), filepath.Join(destinationFolderPath, conceptSchemeVersion.ID+"_for_dspace.xml"))
	if err != nil {
		zapLogger.Debug(err.Error())
		return err
	}
	_, err = copyFile(conceptSchemeVersion.WorkingFilePathNTriples, filepath.Join(destinationFolderPath, conceptSchemeVersion.ID+".nt"))
	if err != nil {
		zapLogger.Debug(err.Error())
		return err
	}
	return err
}

//func (website *Website) GenerateZip(conceptSchemeVersion *ConceptSchemeVersion) error {
//	filesPathsToZip := []string{conceptSchemeVersion.WorkingFilePathNTriples, filepath.Join(conceptSchemeVersion.SkosProcessedFolderPath, conceptSchemeVersion.ID+"_for_dspace.xml")}
//	err := zipFiles(filepath.Join(website.StaticContentFolderPath, fmt.Sprint(conceptSchemeVersion.ID, "_", conceptSchemeVersion.Version, ".zip")), filesPathsToZip)
//	if err != nil {
//		zapLogger.Debug(err.Error())
//		return err
//	}
//	return err
//}

func (website *Website) GenerateHtmlTree(conceptSchemeVersion *ConceptSchemeVersion, asCurrentVersion bool) error {
	var err error
	html := "<ul id=\"tree-root\">"
	treeDepth := 0
	finalNodeDepth := 0
	var treeFunction ki.Func
	treeFunction = func(k ki.Ki, level int, data interface{}) bool {
		concept := conceptSchemeVersion.GetConceptById(k.Name())
		finalNodeDepth = level
		if level > treeDepth {
			html += "<ul>"
		} else if level < treeDepth {
			stepsBack := treeDepth - level
			for i := 1; i <= stepsBack; i++ {
				html += "</ul>"
			}
		}
		treeDepth = level
		if k.Name() != conceptSchemeVersion.ID {
			if concept.Deprecated != true {
				html += fmt.Sprintf("<li><a href=\"/%s/%s/\">%s</a></li>", conceptSchemeVersion.ID, concept.ID, concept.Title)
			}
			//if asCurrentVersion {
			//	html += fmt.Sprintf("<li><a href=\"/%s/%s/\">%s</a></li>", conceptSchemeVersion.ID, concept.ID, concept.Title)
			//} else {
			//	html += fmt.Sprintf("<li><a href=\"/%s/%s/%s/\">%s</a></li>", conceptSchemeVersion.ID, conceptSchemeVersion.Version, concept.ID, concept.Title)
			//}
		}
		return true
	}
	conceptSchemeVersion.FuncDownMeFirst(treeDepth, conceptSchemeVersion, treeFunction)
	for i := 0; i <= finalNodeDepth; i++ {
		html += "</ul>"
	}
	err = ioutil.WriteFile(filepath.Join(conceptSchemeVersion.CalculateFolderPath(website.ContentFolderPath, asCurrentVersion), "tree.html"), []byte(html), os.ModePerm)
	if err != nil {
		zapLogger.Error(err.Error())
		return err
	}
	return err
}
