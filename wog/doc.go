// Package wog is the Worldographer adapter: it reads and writes .wxx files
// to and from the ottomap domain model.
//
// .wxx is a gzipped, UTF-16/BE XML document. The wog package handles that
// pipeline; callers see only ottomap.Map values.
//
// Reading is version-agnostic: wog.Read sniffs the file's release version
// and dispatches to the appropriate schema decoder. The caller does not need
// to know what version the file is.
//
// Writing requires the caller to choose a target schema version via
// wog.WriteOptions.Version. Callers that need to target multiple versions
// call wog.Write multiple times with different options.
package wog
