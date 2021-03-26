package main

import (
	"github.com/goki/ki/ki"
	"github.com/knakk/rdf"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"sort"
	"strings"
	"time"
)

type Concept struct {
	ki.Node          `yaml:"-"`
	ID               string    `yaml:"-"`
	Title            string    `yaml:"title"`
	Uri              string    `yaml:"uri"`
	Definition       string    `yaml:"description"`
	Deprecated       bool      `yaml:"deprecated"`
	HugoLayout       string    `yaml:"layout"`
	EditorialNote    string    `yaml:"-"`
	RelatedMatches   []Match   `yaml:"related"`
	PrefLabels       []Label   `yaml:"pref_labels"`
	AltLabels        []Label   `yaml:"alt_labels"`
	Body             string    `yaml:"-"`
	Updated          time.Time `yaml:"date"`
	IsTopConcept     bool      `yaml:"isTopConcept"`
	NarrowerConcepts []string  `yaml:"narrower_concepts"`
	BroaderConcepts  []string  `yaml:"broader_concepts"`
}

func (concept *Concept) initialise(conceptUri, namespace, conceptSchemeUri string, triplesForThisConcept []rdf.Triple, updated time.Time) {
	concept.Uri = conceptUri
	concept.Updated = updated
	concept.HugoLayout = "concept"
	concept.ID = getConceptNameFromUri(namespace, conceptUri)
	for _, skosConcept := range triplesForThisConcept {
		if skosConcept.Pred.String() == "http://www.w3.org/2004/02/skos/core#prefLabel" {
			languageTag := languageTagFromLiteral(skosConcept.Obj.Serialize(rdf.NTriples))
			if languageTag == "en" {
				concept.Title = strings.TrimSpace(skosConcept.Obj.String())
			}
			concept.PrefLabels = append(concept.PrefLabels, Label{languageTag, languageNameFromTag(languageTag), skosConcept.Obj.String()})
		}
		if skosConcept.Pred.String() == "http://www.w3.org/2004/02/skos/core#altLabel" {
			languageTag := languageTagFromLiteral(skosConcept.Obj.Serialize(rdf.NTriples))
			concept.AltLabels = append(concept.AltLabels, Label{languageTag, languageNameFromTag(languageTag), skosConcept.Obj.String()})
		}
		if skosConcept.Pred.String() == "http://www.w3.org/2004/02/skos/core#exactMatch" {
			concept.RelatedMatches = append(concept.RelatedMatches, Match{MatchType: "Exact Match", MatchUri: skosConcept.Obj.String()})
		}
		if skosConcept.Pred.String() == "http://www.w3.org/2004/02/skos/core#broadMatch" {
			concept.RelatedMatches = append(concept.RelatedMatches, Match{MatchType: "Broad Match", MatchUri: skosConcept.Obj.String()})
		}
		if skosConcept.Pred.String() == "http://www.w3.org/2004/02/skos/core#narrowMatch" {
			concept.RelatedMatches = append(concept.RelatedMatches, Match{MatchType: "Narrow Match", MatchUri: skosConcept.Obj.String()})
		}
		if skosConcept.Pred.String() == "http://www.w3.org/2004/02/skos/core#relatedMatch" {
			concept.RelatedMatches = append(concept.RelatedMatches, Match{MatchType: "Related Match", MatchUri: skosConcept.Obj.String()})
		}
		if skosConcept.Pred.String() == "http://www.w3.org/2004/02/skos/core#definition" {
			dDefinitionLanguageTag := languageTagFromLiteral(skosConcept.Obj.Serialize(rdf.NTriples))
			if dDefinitionLanguageTag == "en" {
				concept.Definition = strings.TrimSpace(skosConcept.Obj.String())
			}
		}
		if skosConcept.Pred.String() == "http://www.w3.org/2004/02/skos/core#topConceptOf" && skosConcept.Obj.String() == conceptSchemeUri {
			concept.IsTopConcept = true
		}
		if skosConcept.Pred.String() == (namespace + "schema#expires") {
			zapLogger.Info("Deprecating", zap.String("concept ID", concept.ID))
			concept.Deprecated = true
		}
		if skosConcept.Pred.String() == "http://www.w3.org/2004/02/skos/core#narrower" {
			concept.NarrowerConcepts = append(concept.NarrowerConcepts, getConceptNameFromUri(namespace, skosConcept.Obj.String()))
		}
		if skosConcept.Pred.String() == "http://www.w3.org/2004/02/skos/core#broader" {
			concept.BroaderConcepts = append(concept.BroaderConcepts, getConceptNameFromUri(namespace, skosConcept.Obj.String()))
		}
	}
	// Sort the various slices
	sort.Slice(concept.PrefLabels, func(i, j int) bool {
		return concept.PrefLabels[i].Value < concept.PrefLabels[j].Value
	})
	sort.Slice(concept.AltLabels, func(i, j int) bool {
		return concept.AltLabels[i].Value < concept.AltLabels[j].Value
	})
	sort.Slice(concept.RelatedMatches, func(i, j int) bool {
		return concept.RelatedMatches[i].MatchType < concept.RelatedMatches[j].MatchType
	})
}

func (concept *Concept) marshal() ([]byte, error) {
	webpageBytes, err := yaml.Marshal(concept)
	finalPage := append([]byte("---\n"), webpageBytes...)
	finalPage = append(finalPage, []byte("---\n\n")...)
	finalPage = append(finalPage, concept.Body...)
	return finalPage, err
}
