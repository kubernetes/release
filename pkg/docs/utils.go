/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package docs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type PrData struct {
	KubernetesPR int
	KEP          int
	WebsitePR    int
	Assignee     string
	Sig          string
}

func StructureData(data [][]interface{}, index []int) ([]*PrData, error) {
	if len(index) < 5 {
		return nil, fmt.Errorf("index cannot be less than five")
	}

	prData := []*PrData{}

	for _, row := range data {
		var newPrData PrData
		kepIdx, kwIndx, sigIdx, asgnIdx, kkIdx := index[0], index[1], index[2], index[3], index[4]
		lastIdx := len(row) - 1

		if lastIdx >= kepIdx {
			kepURL := fmt.Sprintf("%s", row[kepIdx])
			id, err := getIDFromURL(kepURL)
			if err != nil {
				logrus.Errorf("unable to get kep id from sheet. err: %v", err)
			} else {
				newPrData.KEP = id
			}
		}

		if lastIdx >= kwIndx {
			kwURL := fmt.Sprintf("%s", row[kwIndx])
			id, err := getIDFromURL(kwURL)
			if err != nil {
				logrus.Errorf("unable to get kep id from sheet. err: %v", err)
			} else {
				newPrData.WebsitePR = id
			}
		}

		if lastIdx >= sigIdx {
			sig := fmt.Sprintf("%s", row[sigIdx])
			newPrData.Sig = sig
		}

		if lastIdx >= asgnIdx {
			asgn := fmt.Sprintf("%s", row[asgnIdx])
			newPrData.Assignee = asgn
		}

		if lastIdx >= kkIdx {
			kkURL := fmt.Sprintf("%s", row[kkIdx])
			id, err := getIDFromURL(kkURL)
			if err != nil {
				logrus.Errorf("unable to get k/k pr id from sheet. err: %v", err)
			} else {
				newPrData.KubernetesPR = id
			}
		}

		prData = append(prData, &newPrData)
	}

	return prData, nil
}

func getIDFromURL(url string) (int, error) {
	if url == "" {
		return 0, fmt.Errorf("can't get id for empty string")
	}

	strArr := strings.Split(url, "/")
	strID := strArr[len(strArr)-1]
	id, err := strconv.Atoi(strID)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func toInt(s string) (int, error) {
	if len(s) != 1 {
		return 0, fmt.Errorf("column header must be one uppercase letter")
	}
	s = strings.ToUpper(s)
	// Subtract from 65 so that uppercase A starts from 0
	return int(rune(s[0])) - 65, nil
}

func ConvertColumnLettersToInt(headers []string) ([]int, error) {
	ret := []int{}

	for _, header := range headers {
		headerInt, err := toInt(header)
		if err != nil {
			return nil, err
		}

		ret = append(ret, headerInt)
	}

	return ret, nil
}
