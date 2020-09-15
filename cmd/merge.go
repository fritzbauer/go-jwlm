package cmd

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/AndreasSko/go-jwlm/merger"
	"github.com/AndreasSko/go-jwlm/model"
	"github.com/jedib0t/go-pretty/table"
	"github.com/spf13/cobra"

	log "github.com/sirupsen/logrus"
)

// mergeCmd represents the merge command
var mergeCmd = &cobra.Command{
	Use:   "merge <left-backup> <right-backup> <dest-filename>",
	Short: "Merge two JW Library backup files",
	Long: `merge imports the left and right .jwlibrary backup file, merges them and 
exports it to the destination file. If a collision between the left and 
the right backup is detected, the user is asked to choose which side should
be included in the merged backup.`,
	Example: "go-jwlm left.jwlibrary right.jwlibrary merged.jwlibrary",
	Run: func(cmd *cobra.Command, args []string) {
		leftFilename := args[0]
		rightFilename := args[1]
		mergedFilename := args[2]
		merge(leftFilename, rightFilename, mergedFilename)
	},
	Args: cobra.ExactArgs(3),
}

func merge(leftFilename string, rightFilename string, mergedFilename string) {
	log.Info("Importing left backup")
	left := model.Database{}
	err := left.ImportJWLBackup(leftFilename)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Importing right backup")
	right := model.Database{}
	err = right.ImportJWLBackup(rightFilename)
	if err != nil {
		log.Fatal(err)
	}

	merged := model.Database{}

	log.Info("Merging Locations")
	mergedLocations, locationIDChanges, err := merger.MergeLocations(left.Location, right.Location)
	merged.Location = mergedLocations
	merger.UpdateIDs(left.Bookmark, right.Bookmark, "LocationID", locationIDChanges)
	merger.UpdateIDs(left.Bookmark, right.Bookmark, "PublicationLocationID", locationIDChanges)
	merger.UpdateIDs(left.Note, right.Note, "LocationID", locationIDChanges)
	merger.UpdateIDs(left.TagMap, right.TagMap, "LocationID", locationIDChanges)

	log.Info("Merging Bookmarks")
	var bookmarksConflictSolution map[string]merger.MergeSolution
	for {
		mergedBookmarks, _, err := merger.MergeBookmarks(left.Bookmark, right.Bookmark, bookmarksConflictSolution)
		if err == nil {
			merged.Bookmark = mergedBookmarks
			break
		}
		switch err := err.(type) {
		case merger.MergeConflictError:
			bookmarksConflictSolution = handleMergeConflict(err.Conflicts, &left, &right)
		default:
			log.Fatal(err)
		}
	}

	log.Info("Merging Tags")
	var tagsConflictSolution map[string]merger.MergeSolution
	for {
		mergedTags, tagIDChanges, err := merger.MergeTags(left.Tag, right.Tag, tagsConflictSolution)
		if err == nil {
			merged.Tag = mergedTags
			merger.UpdateIDs(left.TagMap, right.TagMap, "TagID", tagIDChanges)
			break
		}
		switch err := err.(type) {
		case merger.MergeConflictError:
			tagsConflictSolution = handleMergeConflict(err.Conflicts, &left, &right)
		default:
			log.Fatal(err)
		}
	}

	log.Info("Merging TagMaps")
	var tagMapsConflictSolution map[string]merger.MergeSolution
	for {
		mergedTagMaps, _, err := merger.MergeTagMaps(left.TagMap, right.TagMap, tagMapsConflictSolution)
		if err == nil {
			merged.TagMap = mergedTagMaps
			break
		}
		switch err := err.(type) {
		case merger.MergeConflictError:
			tagMapsConflictSolution = handleMergeConflict(err.Conflicts, &left, &right)
		default:
			log.Fatal(err)
		}
	}

	log.Info("Merging UserMarks & BlockRanges")
	var UMBRConflictSolution map[string]merger.MergeSolution
	for {
		mergedUserMarks, mergedBlockRanges, userMarkIDChanges, err := merger.MergeUserMarkAndBlockRange(left.UserMark, left.BlockRange, right.UserMark, right.BlockRange, UMBRConflictSolution)
		if err == nil {
			merged.UserMark = mergedUserMarks
			merged.BlockRange = mergedBlockRanges
			merger.UpdateIDs(left.Note, right.Note, "UserMarkID", userMarkIDChanges)
			break
		}
		switch err := err.(type) {
		case merger.MergeConflictError:
			UMBRConflictSolution = handleMergeConflict(err.Conflicts, &left, &right)
		default:
			log.Fatal(err)
		}
	}

	log.Info("Merging Notes")
	var notesConflictSolution map[string]merger.MergeSolution
	for {
		mergedNotes, notesIDChanges, err := merger.MergeNotes(left.Note, right.Note, notesConflictSolution)
		if err == nil {
			merged.Note = mergedNotes
			merger.UpdateIDs(merged.TagMap, nil, "NoteID", notesIDChanges)
			break
		}
		switch err := err.(type) {
		case merger.MergeConflictError:
			notesConflictSolution = handleMergeConflict(err.Conflicts, &left, &right)
		default:
			log.Fatal(err)
		}
	}

	log.Info("Exporting merged database")
	if err = merged.ExportJWLBackup(mergedFilename); err != nil {
		log.Fatal(err)
	}

}

func handleMergeConflict(conflicts map[string]merger.MergeConflict, leftDB *model.Database, rightDB *model.Database) map[string]merger.MergeSolution {
	prompt := &survey.Select{
		Message: "Select which side should be chosen:",
		Options: []string{"Left", "Right"},
	}

	result := make(map[string]merger.MergeSolution, len(conflicts))
	for key, conflict := range conflicts {
		t := table.NewWriter()
		t.SetStyle(table.StyleRounded)

		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Left", "Right"})
		t.AppendRow([]interface{}{conflict.Left.PrettyPrint(leftDB), conflict.Right.PrettyPrint(rightDB)})
		t.Render()

		fmt.Print("\n\n")

		var selected string
		err := survey.AskOne(prompt, &selected)
		if err == terminal.InterruptErr {
			fmt.Println("interrupted")
			os.Exit(0)
		} else if err != nil {
			panic(err)
		}

		if selected == "Left" {
			result[key] = merger.MergeSolution{
				Side:      merger.LeftSide,
				Solution:  conflict.Left,
				Discarded: conflict.Right,
			}
		} else {
			result[key] = merger.MergeSolution{
				Side:      merger.RightSide,
				Solution:  conflict.Right,
				Discarded: conflict.Left,
			}
		}
	}

	return result
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}