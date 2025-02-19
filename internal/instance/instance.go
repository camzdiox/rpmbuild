package instance

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/geteduroam/linux-app/internal/utils"
)

type geo struct {
	Lat  float32 `json:"lat"`
	Long float32 `json:"long"`
}

type Instance struct {
	CatIDP   int       `json:"cat_idp"`
	Country  string    `json:"country"`
	Geo      []geo     `json:"geo"`
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Profiles []Profile `json:"profiles"`
}

type Instances []Instance

func SortNames(a string, b string, search string) int {
	la := strings.ToLower(a)
	lb := strings.ToLower(b)
	bd := strings.Compare(la, lb)
	// compute the base difference which is based on alphabetical order
	// if no search is defined return the base difference
	if search == "" {
		return bd
	}
	match := regexp.MustCompile(fmt.Sprintf("(^|[\\P{L}])%s[\\P{L}]", strings.ToLower(search)))
	mi := match.MatchString(la)
	mj := match.MatchString(lb)
	if mi == mj {
		// tiebreak on alphabetical order
		return bd
	} else if mi {
		return -1
	}
	return 1
}

type ByName struct {
	Instances Instances
	Search    string
}

func (s ByName) Len() int      { return len(s.Instances) }
func (s ByName) Swap(i, j int) { s.Instances[i], s.Instances[j] = s.Instances[j], s.Instances[i] }
func (s ByName) Less(i, j int) bool {
	diff := SortNames(s.Instances[i].Name, s.Instances[j].Name, s.Search)
	// if i is less than j, diff returns less than 0
	return diff < 0
}

func FilterSingle(name string, search string) bool {
	l1, err1 := utils.RemoveDiacritics(strings.ToLower(name))
	l2, err2 := utils.RemoveDiacritics(strings.ToLower(search))
	if err1 != nil || err2 != nil {
		return false
	}
	if !strings.Contains(l1, l2) {
		return false
	}
	return true
}

// FilterSort filters and sorts a list of instances
// The sorting is done in reverse as this is used in the CLI where the most relevant instances should be shown at the bottom
func (i *Instances) FilterSort(search string) *Instances {
	x := ByName{
		Instances: Instances{},
		Search:    search,
	}
	for _, i := range *i {
		if FilterSingle(i.Name, search) {
			x.Instances = append(x.Instances, i)
		}
	}
	sort.Sort(sort.Reverse(ByName(x)))
	return &x.Instances
}
