# COAR Vocabularies-website Processor
Tool for generating the COAR Vocabularies documentation from SKOS files, exported from the COAR Vocabularies Management system (iQVoc).

This tool generates sources designed to work with Hugo. However, they could probably be quite easilty ported to any static site generator that uses Markdown with YAML metadata front-matter.



## Pre-requisites

Source skos files must be in `ntriples` format
Source SKOS files *must* contain *one* `conceptScheme`

Recommended to run the SKOS files through Skosify, to clean up the syntax and optionally add some semantics. Also, exports from iQVoc lack a conceptScheme, and Skosify can automatically add one of these.

```bash
skosify \
  --namespace="http://purl.org/coar/access_right/" \
  --label="Access Rights" \
  --set-modified \
  --mark-top-concepts \
  --narrower \
  --eliminate-redundancy \
  --cleanup-classes \
  --cleanup-properties \
  --cleanup-unreachable \
  --no-enrich-mappings \
  -o ./concept_scheme_sources/access_rights/1.0/concept_scheme.nt \
  ~/_temp/downloaded_concept_schemes/access_rights/1.0/concept_scheme.nt
```

```bash
skosify \
  --namespace="http://purl.org/coar/resource_type/" \
  --label="Resource Types" \
  --set-modified \
  --mark-top-concepts \
  --narrower \
  --eliminate-redundancy \
  --cleanup-classes \
  --cleanup-properties \
  --cleanup-unreachable \
  --no-enrich-mappings \
  -o ./concept_scheme_sources/resource_types/3.0/concept_scheme.nt \
  ~/_temp/downloaded_concept_schemes/resource_types/3.0/concept_scheme.nt
```

```bash
skosify \
  --namespace="http://purl.org/coar/version/" \
  --label="Version Types" \
  --set-modified \
  --mark-top-concepts \
  --narrower \
  --eliminate-redundancy \
  --cleanup-classes \
  --cleanup-properties \
  --cleanup-unreachable \
  --no-enrich-mappings \
  -o ./concept_scheme_sources/version_types/1.0/concept_scheme.nt \
  ~/_temp/downloaded_concept_schemes/version_types/1.0/concept_scheme.nt
```

## Running the processor
1. Prepare the SKOS input files (see above)
2. Compile the Go code in `./src`
3. Copy the `config/config_TEMPLATE.yaml` file to `config/config.yaml`
4. Enter the path to a valid Hugo website source folder in the `webroot` property in `config/config.yaml`
5. Assuming that the compiled binary is in `./binaries`, from within this root directory run: `./binaries/coar_website_builder -c ./config/config.yaml`