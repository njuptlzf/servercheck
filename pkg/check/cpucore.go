package check

import (
	"fmt"
	"runtime"

	v1 "github.com/njuptlzf/servercheck/api/check/v1"
	// "github.com/njuptlzf/servercheck/pkg/option"
	"github.com/juju/errors"
	"github.com/njuptlzf/servercheck/pkg/register"
)

type CPUCoreChecker struct {
	// Name
	name string
	// Specific check item
	item *CPUCoreOption
	// Detailed description
	description string
	// Suggestion on failure
	suggestionOnFail string
	// Return code: fail, warn, ok
	rc v1.ReturnCode
	// Actual check result
	result string
	// Dedicated retrieval interface
	retriever CPURetriever
}

type CPUCoreOption struct {
	// Number of cores
	number int
	// todo: to support
	// cpu usage
	// usage float64
}

var _ v1.Checker = &CPUCoreChecker{}

func init() {
	register.RegisterCheck(newCPUChecker(&CPUCoreOption{
		number: 4,
	}, &RealCPURetriever{}))
}

type CPURetriever interface {
	Get() (*CPUCoreOption, error)
}

type RealCPURetriever struct{}

var _ CPURetriever = &RealCPURetriever{}

func (r *RealCPURetriever) Get() (*CPUCoreOption, error) {
	return &CPUCoreOption{
		number: runtime.NumCPU(),
	}, nil
}

func newCPUChecker(item *CPUCoreOption, retriever CPURetriever) *CPUCoreChecker {
	return &CPUCoreChecker{
		name:        "CPUCore",
		description: "check CPU core",
		item:        item,
		retriever:   retriever,
	}
}

func (c *CPUCoreChecker) Check() error {
	actual, err := c.retriever.Get()
	if err != nil {
		return errors.Trace(err)
	}

	if c.diff(actual) {
		c.rc = v1.PASS
	} else {
		c.rc = v1.WARN
	}
	return nil
}

func (c *CPUCoreChecker) diff(actual *CPUCoreOption) bool {
	pass := true
	coreNumInfo := fmt.Sprintf("[number of cores] acutal: %d, expect: %d", actual.number, c.item.number)
	c.result += coreNumInfo

	if actual.number < c.item.number {
		pass = false
		c.suggestionOnFail += "[number of cores] increase server's CPU"
	}
	return pass
}

func (c *CPUCoreChecker) Name() string {
	return c.name
}

func (c *CPUCoreChecker) Description() string {
	return c.description
}

func (c *CPUCoreChecker) ReturnCode() v1.ReturnCode {
	return c.rc
}

func (c *CPUCoreChecker) Result() string {
	return c.result
}

func (c *CPUCoreChecker) SuggestionOnFail() string {
	return c.suggestionOnFail
}