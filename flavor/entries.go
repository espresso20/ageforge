package flavor

import (
	"io/fs"
	"strings"
)

// Entry is one text of the catalog files as the copy tool reads it: a sentence
// of a pool, or an entry of a noun bank.
type Entry struct {
	File     string // the file it is written in, "exp_fail.txt"
	Line     int    // its line in that file, from 1
	From, To int    // the byte offsets of its text in that file
	ID       string // its row id, "flavor.expedition_fail.any.001"
	Text     string // a sentence with ~ for its slot and no full stop, or a bank entry
	Pool     string // the pool's name ("exp_fail_any") or the bank's ("exp_fail_kit")
	Bucket   string // for a sentence, the pool's part of its row id
	Index    int    // its place in its pool or bank, from 0
	Size     int    // how many entries its pool or bank has
	Bank     bool   // an entry of a noun bank
	Eras     []string
	Ages     []string
	Kinds    []string
	Tones    []string
	Slot     string
}

// ReadEntries parses the catalog files in folder and lists every entry, file
// by file in the order of the copy export, each file's banks and then its
// pools. The game reads the same files from its own build; the copy tool reads
// them from the source tree, so a rewrite it has just made shows in the next
// export.
func ReadEntries(folder fs.FS) ([]Entry, error) {
	catalog, err := readCatalogText(folder)
	if err != nil {
		return nil, err
	}
	var out []Entry
	for _, file := range catalog.files {
		out = append(out, file.entries()...)
	}
	return out, nil
}

func (f fileText) entries() []Entry {
	var out []Entry
	for _, bank := range f.banks {
		for i, word := range bank.words {
			out = append(out, Entry{
				File: f.name + ".txt", Line: word.Line, From: word.Start, To: word.End,
				ID: wordRow(bank.name, i), Text: word.Text, Pool: bank.name,
				Index: i, Size: len(bank.words), Bank: true,
			})
		}
	}
	for _, pool := range f.pools {
		key, _ := rowParts(pool.name)
		for i, sentence := range pool.sentences {
			out = append(out, Entry{
				File: f.name + ".txt", Line: sentence.line, From: sentence.from, To: sentence.to,
				ID: sentenceRow(key, pool.bucket, i), Text: sentence.Text, Pool: pool.name,
				Bucket: pool.bucket, Index: i, Size: len(pool.sentences),
				Eras: eraTextNames(pool.eras), Ages: pool.ages,
				Kinds: sentence.Kinds, Tones: toneTextNames(sentence.Tones), Slot: sentence.Slot,
			})
		}
	}
	return out
}

func eraTextNames(eras []era) []string {
	var out []string
	for _, which := range eras {
		out = append(out, eraTextName(which))
	}
	return out
}

func toneTextNames(tones []Tone) []string {
	var out []string
	for _, tone := range tones {
		out = append(out, strings.ToLower(tone.String()))
	}
	return out
}
