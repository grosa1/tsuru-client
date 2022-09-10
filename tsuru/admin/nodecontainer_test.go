// Copyright 2015 tsuru authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package admin

import (
	"bytes"
	"net/http"

	"github.com/tsuru/tsuru/cmd"
	"github.com/tsuru/tsuru/cmd/cmdtest"
	"gopkg.in/check.v1"
)

func (s *S) TestNodeContainerInfoInfo(c *check.C) {
	c.Assert((&NodeContainerInfo{}).Info(), check.NotNil)
}
func (s *S) TestNodeContainerInfoRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"n1"}, Stdout: &buf}
	body := `{"": {"config": {"image": "img1"}}, "p1": {"config": {"image": "img2"}}, "p2": {"disabled": false, "config": {"image": "img2"}}}`
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: body, Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/1.0/docker/nodecontainers/n1" && req.Method == "GET"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	command := NodeContainerInfo{}
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, `+-------+--------------------------+
| Pool  | Config                   |
+-------+--------------------------+
| <all> | {                        |
|       |   "Name": "",            |
|       |   "PinnedImage": "",     |
|       |   "Disabled": null,      |
|       |   "Config": {            |
|       |     "Cmd": null,         |
|       |     "Image": "img1",     |
|       |     "Entrypoint": null   |
|       |   },                     |
|       |   "HostConfig": {        |
|       |     "ConsoleSize": [     |
|       |       0,                 |
|       |       0                  |
|       |     ],                   |
|       |     "RestartPolicy": {}, |
|       |     "LogConfig": {}      |
|       |   }                      |
|       | }                        |
+-------+--------------------------+
| p1    | {                        |
|       |   "Name": "",            |
|       |   "PinnedImage": "",     |
|       |   "Disabled": null,      |
|       |   "Config": {            |
|       |     "Cmd": null,         |
|       |     "Image": "img2",     |
|       |     "Entrypoint": null   |
|       |   },                     |
|       |   "HostConfig": {        |
|       |     "ConsoleSize": [     |
|       |       0,                 |
|       |       0                  |
|       |     ],                   |
|       |     "RestartPolicy": {}, |
|       |     "LogConfig": {}      |
|       |   }                      |
|       | }                        |
+-------+--------------------------+
| p2    | {                        |
|       |   "Name": "",            |
|       |   "PinnedImage": "",     |
|       |   "Disabled": false,     |
|       |   "Config": {            |
|       |     "Cmd": null,         |
|       |     "Image": "img2",     |
|       |     "Entrypoint": null   |
|       |   },                     |
|       |   "HostConfig": {        |
|       |     "ConsoleSize": [     |
|       |       0,                 |
|       |       0                  |
|       |     ],                   |
|       |     "RestartPolicy": {}, |
|       |     "LogConfig": {}      |
|       |   }                      |
|       | }                        |
+-------+--------------------------+
`)
}

func (s *S) TestNodeContainerListInfo(c *check.C) {
	c.Assert((&NodeContainerList{}).Info(), check.NotNil)
}

func (s *S) TestNodeContainerListRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{}, Stdout: &buf}
	body := `[
{"name": "big-sibling", "configpools": {"": {"config": {"image": "img1"}}, "p1": {"config": {"image": "img2"}}}},
{"name": "c2", "configpools": {"p2": {"config": {"image": "imgX"}}}}
]`
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: body, Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/1.0/docker/nodecontainers" && req.Method == "GET"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	command := NodeContainerList{}
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, `+-------------+--------------+-------+
| Name        | Pool Configs | Image |
+-------------+--------------+-------+
| big-sibling | <all>        | img1  |
|             | p1           | img2  |
+-------------+--------------+-------+
| c2          | p2           | imgX  |
+-------------+--------------+-------+
`)
}

func (s *S) TestNodeContainerAddInfo(c *check.C) {
	c.Assert((&NodeContainerAdd{}).Info(), check.NotNil)
}
func (s *S) TestNodeContainerAddRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"n1"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			err := req.ParseForm()
			c.Assert(err, check.IsNil)
			c.Assert(req.FormValue("name"), check.Equals, "n1")
			c.Assert(req.FormValue("config.image"), check.Equals, "img2")
			c.Assert(req.FormValue("hostconfig.binds.0"), check.Equals, "/a:/b")
			c.Assert(req.FormValue("hostconfig.binds.1"), check.Equals, "/c:/d")
			c.Assert(req.FormValue("hostconfig.logconfig.config.a"), check.Equals, "b")
			c.Assert(req.Form["config.exposedports.8080/tcp"], check.DeepEquals, []string{""})
			c.Assert(req.Form["hostconfig.portbindings.8080/tcp.0.hostport"], check.DeepEquals, []string{"80"})
			c.Assert(req.Form["hostconfig.portbindings.8080/tcp.0.hostip"], check.DeepEquals, []string{""})
			return req.URL.Path == "/1.0/docker/nodecontainers" && req.Method == "POST"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	command := NodeContainerAdd{}
	command.Flags().Parse(true, []string{"--image", "img1", "-p", "80:8080", "-v", "/a:/b", "-v", "/c:/d", "-r", "Config.image=img2", "--log-opt", "a=b"})
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, "Node container successfully added.\n")
}

func (s *S) TestNodeContainerUpdateInfo(c *check.C) {
	c.Assert((&NodeContainerUpdate{}).Info(), check.NotNil)
}
func (s *S) TestNodeContainerUpdateRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"n1"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			err := req.ParseForm()
			c.Assert(err, check.IsNil)
			_, isSet := req.Form["config.cmd"]
			c.Assert(isSet, check.Equals, false)
			c.Assert(req.FormValue("disabled"), check.Equals, "")
			c.Assert(req.FormValue("config.cmd.0"), check.Equals, "echo")
			c.Assert(req.FormValue("config.image"), check.Equals, "img2")
			c.Assert(req.FormValue("hostconfig.binds.0"), check.Equals, "/a:/b")
			c.Assert(req.FormValue("hostconfig.binds.1"), check.Equals, "/c:/d")
			c.Assert(req.FormValue("hostconfig.logconfig.config.a"), check.Equals, "b")
			c.Assert(req.Form["config.exposedports.8080/tcp"], check.DeepEquals, []string{""})
			c.Assert(req.Form["hostconfig.portbindings.8080/tcp.0.hostport"], check.DeepEquals, []string{"80"})
			c.Assert(req.Form["hostconfig.portbindings.8080/tcp.0.hostip"], check.DeepEquals, []string{""})
			return req.URL.Path == "/1.0/docker/nodecontainers/n1" && req.Method == "POST"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	command := NodeContainerUpdate{}
	command.Flags().Parse(true, []string{"--image", "img1", "-p", "80:8080", "-v", "/a:/b", "-v", "/c:/d", "-r", "Config.image=img2", "-r", "config.cmd.0=echo", "--log-opt", "a=b"})
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, "Node container successfully updated.\n")
}

func (s *S) TestNodeContainerUpdateRunDisable(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"n1"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			err := req.ParseForm()
			c.Assert(err, check.IsNil)
			_, isSet := req.Form["config.cmd"]
			c.Assert(isSet, check.Equals, false)
			c.Assert(req.FormValue("disabled"), check.Equals, "true")
			c.Assert(req.FormValue("config.cmd.0"), check.Equals, "echo")
			c.Assert(req.FormValue("config.image"), check.Equals, "img2")
			c.Assert(req.FormValue("hostconfig.binds.0"), check.Equals, "/a:/b")
			c.Assert(req.FormValue("hostconfig.binds.1"), check.Equals, "/c:/d")
			c.Assert(req.FormValue("hostconfig.logconfig.config.a"), check.Equals, "b")
			c.Assert(req.Form["config.exposedports.8080/tcp"], check.DeepEquals, []string{""})
			c.Assert(req.Form["hostconfig.portbindings.8080/tcp.0.hostport"], check.DeepEquals, []string{"80"})
			c.Assert(req.Form["hostconfig.portbindings.8080/tcp.0.hostip"], check.DeepEquals, []string{""})
			return req.URL.Path == "/1.0/docker/nodecontainers/n1" && req.Method == "POST"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	command := NodeContainerUpdate{}
	command.Flags().Parse(true, []string{"--disable", "--image", "img1", "-p", "80:8080", "-v", "/a:/b", "-v", "/c:/d", "-r", "Config.image=img2", "-r", "config.cmd.0=echo", "--log-opt", "a=b"})
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, "Node container successfully updated.\n")
}

func (s *S) TestNodeContainerUpdateRunEnable(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"n1"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			err := req.ParseForm()
			c.Assert(err, check.IsNil)
			_, isSet := req.Form["config.cmd"]
			c.Assert(isSet, check.Equals, false)
			c.Assert(req.FormValue("disabled"), check.Equals, "false")
			c.Assert(req.FormValue("config.cmd.0"), check.Equals, "echo")
			c.Assert(req.FormValue("config.image"), check.Equals, "img2")
			c.Assert(req.FormValue("hostconfig.binds.0"), check.Equals, "/a:/b")
			c.Assert(req.FormValue("hostconfig.binds.1"), check.Equals, "/c:/d")
			c.Assert(req.FormValue("hostconfig.logconfig.config.a"), check.Equals, "b")
			c.Assert(req.Form["config.exposedports.8080/tcp"], check.DeepEquals, []string{""})
			c.Assert(req.Form["hostconfig.portbindings.8080/tcp.0.hostport"], check.DeepEquals, []string{"80"})
			c.Assert(req.Form["hostconfig.portbindings.8080/tcp.0.hostip"], check.DeepEquals, []string{""})
			return req.URL.Path == "/1.0/docker/nodecontainers/n1" && req.Method == "POST"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	command := NodeContainerUpdate{}
	command.Flags().Parse(true, []string{"--enable", "--image", "img1", "-p", "80:8080", "-v", "/a:/b", "-v", "/c:/d", "-r", "Config.image=img2", "-r", "config.cmd.0=echo", "--log-opt", "a=b"})
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, "Node container successfully updated.\n")
}

func (s *S) TestNodeContainerDeleteInfo(c *check.C) {
	c.Assert((&NodeContainerDelete{}).Info(), check.NotNil)
}
func (s *S) TestNodeContainerDeleteRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"n1"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			return req.URL.Path == "/1.0/docker/nodecontainers/n1" && req.Method == "DELETE"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	command := NodeContainerDelete{}
	command.Flags().Parse(true, []string{"-y"})
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, "Node container successfully deleted.\n")
}

func (s *S) TestNodeContainerUpgradeInfo(c *check.C) {
	c.Assert((&NodeContainerUpgrade{}).Info(), check.NotNil)
}
func (s *S) TestNodeContainerUpgradeRun(c *check.C) {
	var buf bytes.Buffer
	context := cmd.Context{Args: []string{"n1"}, Stdout: &buf}
	trans := &cmdtest.ConditionalTransport{
		Transport: cmdtest.Transport{Message: "", Status: http.StatusOK},
		CondFunc: func(req *http.Request) bool {
			err := req.ParseForm()
			c.Assert(err, check.IsNil)
			poolName := req.FormValue("pool")
			return req.URL.Path == "/1.0/docker/nodecontainers/n1/upgrade" && req.Method == "POST" && poolName == "theonepool"
		},
	}
	manager := cmd.Manager{}
	client := cmd.NewClient(&http.Client{Transport: trans}, nil, &manager)
	command := NodeContainerUpgrade{}
	command.Flags().Parse(true, []string{"-p", "theonepool", "-y"})
	err := command.Run(&context, client)
	c.Assert(err, check.IsNil)
	c.Assert(buf.String(), check.Equals, "")
}
