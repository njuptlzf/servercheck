package check

import (
	"fmt"

	"github.com/juju/errors"
	v1 "github.com/njuptlzf/servercheck/api/check/v1"
	optionv1 "github.com/njuptlzf/servercheck/api/option/v1"
	"github.com/njuptlzf/servercheck/pkg/option"
	"github.com/njuptlzf/servercheck/pkg/register"
	"github.com/njuptlzf/servercheck/pkg/utils/parse"
	"github.com/njuptlzf/servercheck/pkg/utils/system"
)

var _ v1.Checker = &DiskAvailChecker{}

type DiskAvailChecker struct {
	// Name
	name string
	// Detailed description
	description string
	// Suggestion on failure
	suggestionOnFail string
	// Return code: fail, warn, or ok
	rc v1.ReturnCode
	// Actual check result
	result string
	// Dedicated retrieval interface
	retriever DiskAvailRetriever
}

func init() {
	register.RegisterCheck(newDiskAvailChecker(&RealDiskAvailRetriever{exp: &expDiskAvailOption{Option: option.Opt}}))
}

func newDiskAvailChecker(retriever DiskAvailRetriever) *DiskAvailChecker {
	return &DiskAvailChecker{
		name:        "DiskAvail",
		description: "check DiskAvail",
		retriever:   retriever,
	}
}

func (c *DiskAvailChecker) Check() error {
	exp, act, err := c.retriever.Collect()
	if err != nil {
		return errors.Trace(err)
	}

	// default rc: WARN or FAIL
	c.rc = v1.FAIL
	// c.rc = v1.WARN

	ok, err := c.diff(exp, act)
	if err != nil {
		return errors.Trace(err)
	}
	if ok {
		c.rc = v1.PASS
	}
	return nil
}

func (c *DiskAvailChecker) diff(exp *expDiskAvailOption, act *actDiskAvailOption) (bool, error) {
	pass := true

	// Compare the available space of each directory
	for i, d := range act.diskOfDir {
		// Actual value: parsed into a readable string
		_, actSize, _, err := parse.ParseDiskForDir(d)
		if err != nil {
			return false, errors.Trace(err)
		}
		// Expected value: parsed into a readable string
		dir, expSize, failedSug, err := parse.ParseDiskForDir(exp.DiskOfDir[i])
		if err != nil {
			return false, errors.Trace(err)
		}
		c.result += fmt.Sprintf("%s: act: %s, exp: %s\n", dir, parse.ParseSize(float64(actSize)), parse.ParseSize(float64(expSize)))
		// Compare
		if expSize > actSize {
			pass = false
			c.suggestionOnFail += fmt.Sprintf("%s: %s\n", dir, failedSug)
		}
	}
	c.result = c.result[:len(c.result)-1]
	if !pass {
		c.suggestionOnFail = c.suggestionOnFail[:len(c.suggestionOnFail)-1]
	}
	return pass, nil
}

func (c *DiskAvailChecker) Name() string {
	return c.name
}

func (c *DiskAvailChecker) Description() string {
	return c.description
}

func (c *DiskAvailChecker) ReturnCode() v1.ReturnCode {
	return c.rc
}

func (c *DiskAvailChecker) Result() string {
	return c.result
}

func (c *DiskAvailChecker) SuggestionOnFail() string {
	return c.suggestionOnFail
}

// RealDiskAvailRetriever is a dedicated check item
type RealDiskAvailRetriever struct {
	// expect option value
	exp *expDiskAvailOption

	// actual option value
	act *actDiskAvailOption
}

type expDiskAvailOption struct {
	*optionv1.Option
}

type actDiskAvailOption struct {
	// The minimum available space for each directory, the format is the same as diskOfDir flag
	diskOfDir []string
}

type DiskAvailRetriever interface {
	Collect() (*expDiskAvailOption, *actDiskAvailOption, error)
}

var _ DiskAvailRetriever = &RealDiskAvailRetriever{}

func (r *RealDiskAvailRetriever) Collect() (*expDiskAvailOption, *actDiskAvailOption, error) {
	r.act = &actDiskAvailOption{}
	// The structure of each element: separated by semicolons, directory path, minimum expected value, failed suggestion. For example, /;100G;>= 100G
	// Loop through DiskOfDir, get the actual available space for each directory
	for _, c := range r.exp.DiskOfDir {
		dir, _, _, err := parse.ParseDiskForDir(c)
		if err != nil {
			return nil, nil, errors.Trace(err)
		}
		// Get the actual available space for the directory
		actSize, err := system.GetAvailableSpace(dir)
		if err != nil {
			return nil, nil, errors.Trace(err)
		}
		// Parsed into a readable string
		hunmanSize := parse.ParseSize(actSize)
		r.act.diskOfDir = append(r.act.diskOfDir, fmt.Sprintf("%s;%s;", dir, hunmanSize))
	}
	return r.exp, r.act, nil
}
