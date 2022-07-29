/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tasks

import (
	"encoding/json"

	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/gitlab/models"
	"github.com/apache/incubator-devlake/plugins/helper"
)

type ApiDetailedStatus struct {
	Icon        string
	Text        string
	Label       string
	Group       string
	Tooltip     string
	HasDetails  bool   `json:"has_details"`
	DetailsPath string `json:"details_path"`
	Favicon     string
}

type ApiPipeline struct {
	Id       int `json:"id"`
	Ref      string
	Sha      string
	Status   string
	Tag      bool
	Duration int
	WebUrl   string `json:"web_url"`

	CreatedAt  *helper.Iso8601Time `json:"created_at"`
	UpdatedAt  *helper.Iso8601Time `json:"updated_at"`
	StartedAt  *helper.Iso8601Time `json:"started_at"`
	FinishedAt *helper.Iso8601Time `json:"finished_at"`

	ApiDetailedStatus
}

var ExtractApiPipelinesMeta = core.SubTaskMeta{
	Name:             "extractApiPipelines",
	EntryPoint:       ExtractApiPipelines,
	EnabledByDefault: true,
	Description:      "Extract raw pipelines data into tool layer table GitlabPipeline",
	DomainTypes:      []string{core.DOMAIN_TYPE_CICD},
}

func ExtractApiPipelines(taskCtx core.SubTaskContext) error {
	rawDataSubTaskArgs, data := CreateRawDataSubTaskArgs(taskCtx, RAW_PIPELINE_TABLE)

	extractor, err := helper.NewApiExtractor(helper.ApiExtractorArgs{
		RawDataSubTaskArgs: *rawDataSubTaskArgs,
		Extract: func(row *helper.RawData) ([]interface{}, error) {
			// create gitlab commit
			gitlabApiPipeline := &ApiPipeline{}
			err := json.Unmarshal(row.Data, gitlabApiPipeline)
			if err != nil {
				return nil, err
			}

			duration := int(gitlabApiPipeline.UpdatedAt.ToTime().Sub(gitlabApiPipeline.CreatedAt.ToTime()).Seconds())
			gitlabApiPipeline.Duration = duration
			gitlabPipeline, err := convertPipeline(gitlabApiPipeline, data.Options.ProjectId)
			if err != nil {
				return nil, err
			}

			// use data.Options.ProjectId to set the value of ProjectId for it
			gitlabPipeline.ProjectId = data.Options.ProjectId
			gitlabPipeline.ConnectionId = data.Options.ConnectionId
			results := make([]interface{}, 0, 1)
			results = append(results, gitlabPipeline)

			return results, nil
		},
	})

	if err != nil {
		return err
	}

	return extractor.Execute()
}

func convertPipeline(pipeline *ApiPipeline, projectId int) (*models.GitlabPipeline, error) {
	gitlabPipeline := &models.GitlabPipeline{
		GitlabId:         pipeline.Id,
		ProjectId:        projectId,
		Ref:              pipeline.Ref,
		Sha:              pipeline.Sha,
		WebUrl:           pipeline.WebUrl,
		Status:           pipeline.Status,
		GitlabCreatedAt:  helper.Iso8601TimeToTime(pipeline.CreatedAt),
		GitlabUpdatedAt:  helper.Iso8601TimeToTime(pipeline.UpdatedAt),
		GitlabStartedAt:  helper.Iso8601TimeToTime(pipeline.StartedAt),
		GitlabFinishedAt: helper.Iso8601TimeToTime(pipeline.FinishedAt),
		Duration:         pipeline.Duration,
	}
	return gitlabPipeline, nil
}
