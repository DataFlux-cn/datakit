package jira

import (
	"fmt"
	"time"
	"sync"

	"github.com/andygrunwald/go-jira"

	"gitlab.jiagouyun.com/cloudcare-tools/datakit"
	"gitlab.jiagouyun.com/cloudcare-tools/datakit/io"
)

func (p *JiraParam) active() {
	var queue chan string
	wg := sync.WaitGroup{}

	cnt := 0
	lastQueueId := -1
	issueHasFound := make(map[string]bool)
	ticker := time.NewTicker(time.Duration(20) * time.Minute)
	defer ticker.Stop()

	task := func () {
		issues, err := p.findIssues()
		if err != nil {
			p.log.Errorf("findIssues err: %s", err.Error())
		}
		for _, issue := range issues {
			if _, ok := issueHasFound[issue]; ok {
				continue
			}
			issueHasFound[issue] = true
			cnt += 1
			if cnt/maxIssuesPerQueue != lastQueueId {
				lastQueueId = cnt / maxIssuesPerQueue
				queue = make(chan string, maxIssuesPerQueue)
				wg.Add(1)
				go p.gather(queue, &wg)
			}
			queue <- issue
		}
	}

	task()

	for {
		select {
		case <-ticker.C:
			task()
		case <-datakit.Exit.Wait():
			break
		}
	}

	wg.Wait()
}

func (p *JiraParam) gather(queue chan string, wg *sync.WaitGroup) {
	issueL := make([]string, 0)
	ticker := time.NewTicker(time.Duration(p.input.Interval) * time.Second)
	defer ticker.Stop()

	c, err := p.makeJiraClient()
	if err != nil {
		p.log.Errorf("makeJiraClient err: %s", err.Error())
		return
	}

	for {
		select {
		case issue := <-queue:
			issueL = append(issueL, issue)

		case <-ticker.C:
			for _, issue := range issueL {
				params := *p
				p.input.Issue = issue
				err := params.getMetrics(c)
				if err != nil {
					p.log.Errorf("getMetrics err: %s", err.Error())
				}
			}

		case <-datakit.Exit.Wait():
			wg.Done()
			return
		}
	}
}

func (p *JiraParam) getMetrics(c *jira.Client) error {
	i, resp, err := c.Issue.Get(p.input.Issue, &jira.GetQueryOptions{})
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("http issue get code: %d", resp.StatusCode)
	}
	resp.Body.Close()

	tags := make(map[string]string)
	fields := make(map[string]interface{})

	tags["host"] = p.input.Host
	tags["project_key"] = i.Fields.Project.Key
	tags["project_id"] = i.Fields.Project.ID
	tags["project_name"] = i.Fields.Project.Name

	for tag, tagV := range p.input.Tags {
		tags[tag] = tagV
	}

	fields["id"] = i.ID
	fields["key"] = i.Key
	fields["url"] = i.Self

	if i.Fields != nil {
		fields["type"] = i.Fields.Type.Name
		fields["summary"] = i.Fields.Summary
		if i.Fields.Creator != nil {
			fields["creator"] = i.Fields.Creator.Name
		}
		if i.Fields.Assignee != nil {
			fields["assignee"] = i.Fields.Assignee.Name
		}
		if i.Fields.Reporter != nil {
			fields["reporter"] = i.Fields.Reporter.Name
		}
		if i.Fields.Priority != nil {
			fields["priority"] = i.Fields.Priority.Name
		}
		if i.Fields.Status != nil {
			fields["status"] = i.Fields.Status.Name
		}
	}

	pts, err := io.MakeMetric(p.input.MetricsName, tags, fields, time.Time(i.Fields.Updated))
	if err != nil {
		return err
	}
	p.log.Debug(string(pts))

	err = p.output.IoFeed(pts, io.Metric)
	return err
}

func (p *JiraParam) makeJiraClient() (*jira.Client, error) {
	tp := jira.BasicAuthTransport{
		Username: p.input.Username,
		Password: p.input.Password,
	}
	return jira.NewClient(tp.Client(), p.input.Host)
}

func (t *JiraParam) findIssues() ([]string, error) {
	//指定issue ID
	if t.input.Issue != "" {
		return t.findIssueById()
	}
	//遍历所有项目下所有问题，耗时较长
	if t.input.Project == "" && t.input.Issue == "" {
		return t.findAllIssue()
	}
	//获取指定项目下所有issue
	return t.findIssuesByProject()
}

func (t *JiraParam) findAllIssue() ([]string, error) {
	issueL := make([]string, 0)

	c, err := t.makeJiraClient()
	if err != nil {
		return issueL, err
	}

	ops := jira.SearchOptions{}
	ops.StartAt = 0
	ops.MaxResults = 300
	for {
		select {
		case <-datakit.Exit.Wait():
			return nil, nil
		default:
		}

		cnt := 0
		issues, resp, err := c.Issue.Search("", &ops)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return issueL, nil
		}
		if err != nil {
			return issueL, err
		}

		for _, i := range issues {
			cnt += 1
			issueL = append(issueL, i.ID)
		}
		if cnt != ops.MaxResults {
			break
		}
		ops.StartAt += ops.MaxResults
	}
	return issueL, nil
}

func (t *JiraParam) findIssuesByProject() ([]string, error) {
	issueL := make([]string, 0)

	c, err := t.makeJiraClient()
	if err != nil {
		return issueL, err
	}

	p, resp, err := c.Project.Get(t.input.Project)
	resp.Body.Close()
	if resp.StatusCode != 200 {
		return issueL, nil
	}
	if err != nil {
		return issueL, err
	}
	ops := jira.SearchOptions{}
	ops.StartAt = 0
	ops.MaxResults = 300
	sql := fmt.Sprintf("project=%s", p.Key)

	for {
		select {
		case <-datakit.Exit.Wait():
			return nil, nil
		default:
		}

		issues, _, _ := c.Issue.Search(sql, &ops)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return issueL, nil
		}
		if err != nil {
			return issueL, err
		}

		cnt := 0
		for _, i := range issues {
			cnt += 1
			issueL = append(issueL, i.ID)
		}

		if cnt != ops.MaxResults {
			break
		}
		ops.StartAt += ops.MaxResults
	}
	return issueL, nil
}

func (t *JiraParam) findIssueById() ([]string, error) {
	issueL := make([]string, 0)
	issueL = append(issueL, t.input.Issue)
	return issueL, nil
}