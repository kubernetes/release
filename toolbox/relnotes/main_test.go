package main

import (
	"os"
	"regexp"
	"strconv"
	"testing"

	u "k8s.io/release/toolbox/util"
)

func TestDetermineRange(t *testing.T) {
	tables := []struct {
		owner       string
		repo        string
		branch      string
		branchRange string
		start       string
		end         string
	}{
		{"kubernetes", "kubernetes", "release-1.7", "v1.7.0..v1.7.2", "v1.7.0", "v1.7.2"},
		// TODO: bug fix in original script
		// {"kubernetes", "kubernetes", "release-1.7", "v1.7.20", "v1.7.8", "v1.7.20"},
		{"kubernetes", "kubernetes", "release-1.7", "v1.7.5..", "v1.7.5", "5adaee21de0c5ed1286a00468e09d866605f85f4"},
		{"kubernetes", "kubernetes", "release-1.7", "", "v1.7.8", "5adaee21de0c5ed1286a00468e09d866605f85f4"},
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	c := u.NewClient(githubToken)

	for _, table := range tables {
		s, e, err := determineRange(c, table.owner, table.repo, table.branch, table.branchRange)
		if err != nil {
			t.Errorf("%v %v: Unexpected error: %v", table.branch, table.branchRange, err)
		}
		if s != table.start {
			t.Errorf("%v %v: Start tag was incorrect, want: %s, got: %s", table.branch, table.branchRange, table.start, s)
		}
		if e != table.end {
			t.Errorf("%v %v: End tag was incorrect, want: %s, got: %s", table.branch, table.branchRange, table.end, e)
		}
	}
}

func TestRegExp(t *testing.T) {
	tables := []struct {
		s     string
		cp    bool
		pr    bool
		numCP []int
		numPR int
	}{
		{"Merge pull request #53422 from liggitt/automated-cherry-pick-of-#53233-upstream-release-1.8\n\nAutomatic merge from submit-queue.\n\nAutomated cherry pick of #53233\n\nFixes #51899\r\n\r\nCherry pick of #53233 on release-1.8.\r\n\r\n#53233: remove containers of deleted pods once all containers have\r\n\r\n```release-note\r\nFixes a performance issue (#51899) identified in large-scale clusters when deleting thousands of pods simultaneously across hundreds of nodes, by actively removing containers of deleted pods, rather than waiting for periodic garbage collection and batching resulting pod API deletion requests.\r\n```", true, true, []int{53233}, 53422},

		{"Merge pull request #53448 from liggitt/automated-cherry-pick-of-#53317-upstream-release-1.8\n\nAutomatic merge from submit-queue.\n\nAutomated cherry pick of #53317\n\nCherry pick of #53317 on release-1.8.\n\n#53317: Change default --cert-dir for kubelet to a non-transient", true, true, []int{53317}, 53448},

		{"Merge pull request #53097 from m1093782566/ipvs-test\n\nAutomatic merge from submit-queue (batch tested with PRs 52768, 51898, 53510, 53097, 53058). If you want to cherry-pick this change to another branch, please follow the instructions <a href=\"https://github.com/kubernetes/community/blob/master/contributors/devel/cherry-picks.md\">here</a>.\n\nRun IPVS proxier UTs everywhere - include !linux platfrom\n\n**What this PR does / why we need it**:\r\n\r\nIPVS proxier UTs should run everywhere, including !linux platfrom, which will help a lot when developing in windows platfrom.\r\n\r\n**Which issue this PR fixes**: \r\n\r\nfixes #53099\r\n\r\n**Special notes for your reviewer**:\r\n\r\n**Release note**:\r\n\r\n```release-note\r\nNONE\r\n```", false, true, []int{0}, 53097},

		{"Merge pull request #52602 from liggitt/automated-cherry-pick-of-#48394-#43152-upstream-release-1.7\n\nAutomatic merge from submit-queue.\n\nAutomated cherry pick of #48394 #43152\n\nCherry pick of #48394 #43152 on release-1.7.\r\n\r\n#48394: GuaranteedUpdate must write if stored data is not canonical\r\n#43152: etcd3 store: retry w/live object on conflict", true, true, []int{48394, 43152}, 52602},
	}

	reCherry, _ := regexp.Compile("automated-cherry-pick-of-(#[0-9]+-){1,}")
	reCherryID, _ := regexp.Compile("#([0-9]+)-")
	reMerge, _ := regexp.Compile("^Merge pull request #([0-9]+) from")

	for _, table := range tables {
		// Check cherry pick regexp
		vCP := reCherry.FindStringSubmatch(table.s)

		if table.cp && vCP == nil {
			t.Errorf("Cherry pick message not matched: \n\n%s\n\n", table.s)
		}
		if !table.cp && vCP != nil {
			t.Errorf("Non cherry pick message matched: \n\n%s\n\n", table.s)
		}
		if table.cp && vCP != nil {
			vID := reCherryID.FindAllStringSubmatch(vCP[0], -1)
			if vID == nil {
				t.Errorf("Unexpected empty cherry pick")
			} else {
				if len(vID) != len(table.numCP) {
					t.Errorf("Number of cherry pick PRs mismatch: want: %d, got: %d", len(table.numCP), len(vID))
				}
				for idx, i := range vID {
					id, err := strconv.Atoi(i[1])
					if err != nil {
						t.Errorf("Unexpected error: %v", err)
					} else if table.numCP[idx] != id {
						t.Errorf("Cherry pick id was incorrect, want: %d, got: %d", table.numCP[idx], id)
					}
				}
			}
		}
		// Check normal PR regexp
		vPR := reMerge.FindStringSubmatch(table.s)
		if table.pr && vPR == nil {
			t.Errorf("Normal PR message not matched: \n\n%s\n\n", table.s)
		}
		if !table.pr && vPR != nil {
			t.Errorf("Non normal PR message matched: \n\n%s\n\n", table.s)
		}
		if table.pr && vPR != nil {
			id, err := strconv.Atoi(vPR[1])
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if table.numPR != id {
				t.Errorf("Normal PR id was incorrect, want: %d, got: %d", table.numPR, id)
			}
		}
	}
}

// NOTE: the following tests are for file content generation tests, and require taking looks at
// the generated files to verify the correctness.
func TestGetCIJobStatus(t *testing.T) {
	filename := "/tmp/release_note_CI_status_testfile"
	filenameHTML := "/tmp/release_note_CI_status_testfile_html"

	// Remove existing file, because getCIJobStatus() will append on existing file
	os.Remove(filename)
	os.Remove(filenameHTML)

	err := getCIJobStatus(filename, "release-1.7", false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	err = getCIJobStatus(filenameHTML, "release-1.7", true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestCreateBody(t *testing.T) {
	filename := "/tmp/release_note_body_testfile"
	releaseTag := "v1.7.2"
	branch := "release-1.7"
	docURL := "https://testdoc.com"
	exampleURL := "https://testexample.com"
	releaseTars := "../../../public_kubernetes/_output/release-tars"
	f, err := os.Create(filename)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	createBody(f, releaseTag, branch, docURL, exampleURL, releaseTars)
}

func TestCreateHTMLNote(t *testing.T) {
	htmlFileName := "/tmp/release_note_tests_html_testfile"
	mdFileName := "/tmp/relnotes-release-1.7.md"
	err := createHTMLNote(htmlFileName, mdFileName)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestGetPendingPRs(t *testing.T) {
	filename := "/tmp/release_note_pending_pr_testfile"
	owner := "kubernetes"
	repo := "kubernetes"
	branch := "release-1.7"

	f, err := os.Create(filename)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	c := u.NewClient(githubToken)

	err = getPendingPRs(c, f, owner, repo, branch)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
