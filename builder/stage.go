package builder

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/jinrudals/workscript/builder/utils"
)

type Stage struct {
	Name      string
	Variables map[string]string
	Run       map[string]string
	Post      map[string]string
	Before    []string
	After     []string
	Requires  []string
}

var (
	selfRefPattern  = regexp.MustCompile(`@\{(\w+)\}`)
	crossRefPattern = regexp.MustCompile(`^@\{(.+?)\.(.+?)\}$`)
)

func NewStage(props map[string]interface{}) *Stage {
	stage := &Stage{
		Name:      props["name"].(string),
		Variables: utils.GetStringMap(props["variables"]),
		Run:       utils.GetStringMap(props["run"]),
		Post:      utils.GetStringMap(props["post"]),
		Before:    utils.GetStringSlice(props["before"]),
		After:     utils.GetStringSlice(props["after"]),
		Requires:  utils.GetStringSlice(props["requires"]),
	}
	stage.applyGlobalVariables()
	return stage
}
func (s *Stage) SetName(target string) {
	prefix := target + ":"
	s.Name = prefix + s.Name
	s.Before = utils.PrefixList(prefix, s.Before)
	s.After = utils.PrefixList(prefix, s.After)
	s.Requires = utils.PrefixList(prefix, s.Requires)
}

func (s *Stage) applyGlobalVariables() {
	gv := GetInstance()
	for k, v := range gv.mmap {
		if strings.HasPrefix(k, s.Name+":") {
			envName := strings.TrimPrefix(k, s.Name+":")
			s.Variables[envName] = v
		}
	}
}

func (s *Stage) ResolveSelfVariables() {
	for key, val := range s.Variables {
		s.Variables[key] = selfRefPattern.ReplaceAllStringFunc(val, func(m string) string {
			match := selfRefPattern.FindStringSubmatch(m)
			if len(match) == 2 {
				return s.Variables[match[1]]
			}
			return ""
		})
	}
}

func (s *Stage) ApplyCrossVariables(stages map[string]*Stage) error {
	for key, val := range s.Variables {
		if matches := crossRefPattern.FindStringSubmatch(val); len(matches) == 3 {
			refStageName, refVar := matches[1], matches[2]
			refStage, ok := stages[refStageName]
			if !ok {
				return fmt.Errorf("no such stage: %s", refStageName)
			}
			refVal, ok := refStage.Variables[refVar]
			if !ok {
				return fmt.Errorf("variable '%s' not found in '%s'", refVar, refStageName)
			}
			s.Variables[key] = refVal
			log.Printf("Resolved variable @%s.%s -> %s", refStageName, refVar, refVal)
		}
	}
	return nil
}

func (s *Stage) ResolveCommandVariables() error {
	s.Run = utils.ResolveVars(s.Run, s.Variables, s.Name)
	s.Post = utils.ResolveVars(s.Post, s.Variables, s.Name)
	return nil
}

func (s *Stage) RemoveMissingDependencies(stages map[string]*Stage) {
	s.Before = utils.FilterExisting(s.Before, stages)
	s.After = utils.FilterExisting(s.After, stages)
}

func (s *Stage) ToJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"name":      s.Name,
		"variables": s.Variables,
		"run":       s.Run,
		"post":      s.Post,
		"before":    s.Before,
		"after":     s.After,
		"requires":  s.Requires,
	})
}
