package main

import (
	"fmt"
	"github.com/go-xmlfmt/xmlfmt"
	"github.com/goki/ki/ki"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type ByVersion []ConceptSchemeVersion

func (s ByVersion) Len() int {
	return len(s)
}
func (s ByVersion) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByVersion) Less(i, j int) bool {
	return s[i].Version < s[j].Version
}

type ByReleaseDate []ConceptSchemeVersion

func (s ByReleaseDate) Len() int {
	return len(s)
}
func (s ByReleaseDate) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByReleaseDate) Less(i, j int) bool {
	return s[i].Released.Before(s[j].Released)
}

type ConceptSchemeVersion struct {
	ki.Node                    `yaml:"-"`
	ID                         string              `yaml:"-"`
	Version                    string              `yaml:"version"`
	Current                    bool                `yaml:"current"`
	Latest                     bool                `yaml:"latest"`
	Title                      string              `yaml:"title"`
	Description                string              `yaml:"description"`
	ChangeLog                  string              `yaml:"change_log"`
	Namespace                  string              `yaml:"namespace"`
	Uri                        string              `yaml:"uri"`
	SkosProcessedFolderPath    string              `yaml:"-"`
	WorkingFilePathNTriples    string              `yaml:"-"`
	Released                   time.Time           `yaml:"date"`
	Creators                   []map[string]string `yaml:"creators"`
	Contributors               []string            `yaml:"contributors"`
	Languages                  []string            `yaml:"languages"`
	Concepts                   []*Concept          `yaml:"-"`
	NotDeprecatedConceptIDList []string            `yaml:"not_deprecated_concepts"`
	HugoLayout                 string              `yaml:"layout,omitempty"`
}

func (conceptSchemeVersion *ConceptSchemeVersion) Marshal() ([]byte, error) {
	webpageBytes, err := yaml.Marshal(conceptSchemeVersion)
	finalPage := append([]byte("---\n"), webpageBytes...)
	finalPage = append(finalPage, []byte("---\n\n")...)
	return finalPage, err
}

func (conceptSchemeVersion *ConceptSchemeVersion) GetConceptById(id string) *Concept {
	for _, concept := range conceptSchemeVersion.Concepts {
		if concept.ID == id {
			return concept
		}
	}
	return nil
}

func (conceptSchemeVersion *ConceptSchemeVersion) CalculateFolderPath(webRootContentPath string, asCurrentVersion bool) string {
	if asCurrentVersion {
		return filepath.Join(webRootContentPath, conceptSchemeVersion.ID)
	} else {
		return filepath.Join(webRootContentPath, conceptSchemeVersion.ID, conceptSchemeVersion.Version)
	}
}

func (conceptSchemeVersion *ConceptSchemeVersion) generateDspaceXml() error {
	var err error
	for _, languageCode := range conceptSchemeVersion.Languages {
		err = conceptSchemeVersion.generateDspaceXmlForLanguage(languageCode)
		if err != nil {
			zapLogger.Error(err.Error())
			return err
		}
	}
	return err
}

func (conceptSchemeVersion *ConceptSchemeVersion) generateDspaceXmlForLanguage(languageCode string) error {
	dSpaceXmlFilePath := filepath.Join(conceptSchemeVersion.SkosProcessedFolderPath, conceptSchemeVersion.ID+"_for_dspace_"+languageCode+".xml")
	zapLogger.Debug(fmt.Sprintf("Creating DSpace XML file for '%s: %s' at %s", conceptSchemeVersion.ID, conceptSchemeVersion.Version, dSpaceXmlFilePath))
	var err error
	xml := `
<?xml version="1.0" encoding="UTF-8"?>
<!--This XML was automatically generated from the SKOS sources for the COAR Vocabulary. It follows the schema developed by the University of Minho (https://www.uminho.pt) for DSpace.
UMinho Schema: https://github.com/DSpace/DSpace/blob/main/dspace/config/controlled-vocabularies/controlledvocabulary.xsd
Initial example: https://github.com/4Science/DSpace/blob/dspace-cris-7/dspace/config/controlled-vocabularies/types.xml-->
	`
	currentLevel := 0
	finalNodeDepth := 0
	conceptSchemeVersion.FuncDownMeFirst(0, nil, func(k ki.Ki, level int, d interface{}) bool {
		xmlNode := ""
		concept := conceptSchemeVersion.GetConceptById(k.Name())
		if k.Name() == conceptSchemeVersion.ID {
			xmlNode = fmt.Sprintf("<node id=\"%s\" label=\"%s\">", conceptSchemeVersion.Uri, conceptSchemeVersion.Title)
		} else {
			if concept.Deprecated != true {
				xmlNode = fmt.Sprintf("<node id=\"%s\" label=\"%s\">", concept.ID, concept.Title)
				for _, label := range concept.PrefLabels {
					if label.LanguageCode == languageCode {
						xmlNode = fmt.Sprintf("<node id=\"%s\" label=\"%s\">", concept.ID, label.Value)
					}
				}
				if concept.Definition != "" {
					xmlNode += fmt.Sprintf("<hasNote><![CDATA[%s]]></hasNote>", concept.Definition)
				}
			}
		}
		if level < currentLevel {
			for i := 0; i < (currentLevel - level); i++ {
				xml += "</isComposedBy></node>"
			}
		}
		xml += xmlNode
		if k.HasChildren() {
			xml += "<isComposedBy>"
		} else {
			if concept.Deprecated != true {
				xml += "</node>"
			}
		}
		currentLevel = level
		finalNodeDepth = level
		return true // return value determines whether tree traversal continues or not
	})
	for i := 0; i < finalNodeDepth; i++ {
		xml += "</isComposedBy></node>"
	}
	xml = xmlfmt.FormatXML(xml, "", "  ")
	xml = strings.TrimSpace(xml)
	err = ioutil.WriteFile(dSpaceXmlFilePath, []byte(xml), os.ModePerm)
	if err != nil {
		zapLogger.Error(err.Error())
	}
	return err
}
