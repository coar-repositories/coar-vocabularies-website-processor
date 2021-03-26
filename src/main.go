package main

import (
	"flag"
	"go.uber.org/zap"
)

var zapLogger *zap.Logger
var config = Config{}
var conceptSchemes []*ConceptScheme
var website = Website{}

func main() {
	var err error
	var configFilePath string
	zapLogger, _ = configureZapLogger(false)
	//	### Load arguments
	flag.StringVar(&configFilePath, "c", "", "path to a valid config yaml file")
	flag.Parse()

	// ### Load configuration
	err = (&config).unmarshal(configFilePath)
	if err != nil {
		zapLogger.Fatal("Unable to initialise - halting execution")
	} else {
		zapLogger.Info("Configuration loaded OK")
	}
	// ### Load SKOS configs and initialise conceptSchemes
	conceptSchemes = make([]*ConceptScheme, 0)
	for _, conceptSchemeConfig := range config.ConceptSchemeConfigs {
		var conceptScheme ConceptScheme
		conceptScheme.Initialise(conceptSchemeConfig, config.ProcessedSkosRootFolderPath)
		for _, conceptSchemeVersion := range conceptScheme.Versions {
			err = conceptSchemeVersion.generateDspaceXml()
			if err != nil {
				zapLogger.Fatal("Unable to initialise - halting execution")
			}
		}
		conceptSchemes = append(conceptSchemes, &conceptScheme)
	}
	// ### Initialise Hugo content and static folders
	err = (&website).Initialise(config.WebPageSourcesFolderPath, config.WebrootFolderPath)
	if err != nil {
		zapLogger.Fatal("Unable to initialise website - halting execution")
	}
	zapLogger.Info("Building website...")
	for _, conceptScheme := range conceptSchemes {
		err = website.ProcessConceptScheme(conceptScheme)
		if err != nil {
			zapLogger.Fatal("Unable to process concept scheme for website - halting immediately")
		}
	}
	zapLogger.Info("Website built OK")
	zapLogger.Info("Process completed successfully")
}
