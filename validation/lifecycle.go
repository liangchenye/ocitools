package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	tap "github.com/mndrix/tap-go"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/validation/util"
	uuid "github.com/satori/go.uuid"
)

func main() {
	t := tap.New()
	t.Header(0)

	g := util.GetDefaultGenerator()
	var output string
	config := util.LifecycleConfig{
		Actions: util.LifecycleActionCreate | util.LifecycleActionStart | util.LifecycleActionDelete,
		PreCreate: func(r *util.Runtime) error {
			r.SetID(uuid.NewV4().String())

			g := util.GetDefaultGenerator()
			output = filepath.Join(r.BundleDir, g.Spec().Root.Path, "output")
			prestart := rspec.Hook{
				Path: fmt.Sprintf("%s/%s/bin/sh", r.BundleDir, g.Spec().Root.Path),
				Args: []string{
					"sh", "-c", fmt.Sprintf("echo 'pre-start called' >> %s", output),
				},
			}
			poststart := rspec.Hook{
				Path: fmt.Sprintf("%s/%s/bin/sh", r.BundleDir, g.Spec().Root.Path),
				Args: []string{
					"sh", "-c", fmt.Sprintf("echo 'post-start called' >> %s", output),
				},
			}
			poststop := rspec.Hook{
				Path: fmt.Sprintf("%s/%s/bin/sh", r.BundleDir, g.Spec().Root.Path),
				Args: []string{
					"sh", "-c", fmt.Sprintf("echo 'post-stop called' >> %s", output),
				},
			}

			g.AddPreStartHook(prestart)
			g.AddPostStartHook(poststart)
			g.AddPostStopHook(poststop)
			g.SetProcessArgs([]string{"sh", "-c", fmt.Sprintf("echo 'process called' >> %s", output)})
			r.SetConfig(g)
			return nil
		},
		PostCreate: func(r *util.Runtime) error {
			outputData, err := ioutil.ReadFile(output)
			if err != nil {
				return err
			}
			if string(outputData) != "" {
				return errors.New("Wrong call")
			}
			return nil
		},
		PreDelete: func(r *util.Runtime) error {
			outputData, err := ioutil.ReadFile(output)
			if err != nil {
				return err
			}
			if string(outputData) != "pre-start called\nprocess called\npost-start called\n" {
				return errors.New("Wrong call")
			}
			return nil
		},
		PostDelete: func(r *util.Runtime) error {
			outputData, err := ioutil.ReadFile(output)
			if err != nil {
				return err
			}
			if string(outputData) != "pre-start called\nprocess called\npost-start called\npost-stop called\n" {
				return errors.New("Wrong call")
			}
			return nil
		},
	}

	err := util.RuntimeLifecycleValidate(g, config)
	if err != nil {
		util.Fatal(err)
	}

	t.AutoPlan()
}
