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
	"fmt"
	"github.com/apache/incubator-devlake/plugins/jenkins/models"
	"reflect"
	"time"

	"github.com/apache/incubator-devlake/models/domainlayer"
	"github.com/apache/incubator-devlake/models/domainlayer/devops"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/core/dal"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ConvertBuildsToCICDMeta = core.SubTaskMeta{
	Name:             "convertBuildsToCICD",
	EntryPoint:       ConvertBuildsToCICD,
	EnabledByDefault: true,
	Description:      "convert builds to cicd",
	DomainTypes:      []string{core.DOMAIN_TYPE_CICD},
}

func ConvertBuildsToCICD(taskCtx core.SubTaskContext) error {
	db := taskCtx.GetDal()
	data := taskCtx.GetData().(*JenkinsTaskData)

	clauses := []dal.Clause{
		dal.From("_tool_jenkins_builds"),
		dal.Where("_tool_jenkins_builds.connection_id = ?", data.Options.ConnectionId),
	}
	cursor, err := db.Cursor(clauses...)
	if err != nil {
		return err
	}
	defer cursor.Close()

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		InputRowType: reflect.TypeOf(models.JenkinsBuild{}),
		Input:        cursor,
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Params: JenkinsApiParams{
				ConnectionId: data.Options.ConnectionId,
			},
			Ctx:   taskCtx,
			Table: RAW_BUILD_TABLE,
		},
		Convert: func(inputRow interface{}) ([]interface{}, error) {
			jenkinsBuild := inputRow.(*models.JenkinsBuild)
			durationSec := int64(jenkinsBuild.Duration / 1000)
			jenkinsPipelineResult := ""
			jenkinsPipelineStatus := ""
			var jenkinsPipelineFinishedDate *time.Time
			results := make([]interface{}, 0)
			if jenkinsBuild.Result == "SUCCESS" {
				jenkinsPipelineResult = devops.SUCCESS
			} else if jenkinsBuild.Result == "FAILURE" {
				jenkinsPipelineResult = devops.FAILURE
			} else {
				jenkinsPipelineResult = devops.ABORT
			}

			if jenkinsBuild.Building {
				jenkinsPipelineStatus = devops.IN_PROGRESS
			} else {
				jenkinsPipelineStatus = devops.DONE
				finishTime := jenkinsBuild.StartTime.Add(time.Duration(durationSec * int64(time.Second)))
				jenkinsPipelineFinishedDate = &finishTime
			}
			if jenkinsBuild.TriggeredBy == "" {
				jenkinsPipeline := &devops.CICDPipeline{
					DomainEntity: domainlayer.DomainEntity{
						Id: fmt.Sprintf("%s:%s:%d:%s", "jenkins", "JenkinsPipeline", jenkinsBuild.ConnectionId,
							jenkinsBuild.DisplayName),
					},
					Name:         jenkinsBuild.DisplayName,
					Result:       jenkinsPipelineResult,
					Status:       jenkinsPipelineStatus,
					FinishedDate: jenkinsPipelineFinishedDate,
					Type:         "CI/CD",
					DurationSec:  uint64(durationSec),
					CreatedDate:  jenkinsBuild.StartTime,
				}
				if jenkinsPipelineFinishedDate != nil {
				}
				jenkinsPipeline.RawDataOrigin = jenkinsBuild.RawDataOrigin
				results = append(results, jenkinsPipeline)
			}

			if !jenkinsBuild.HasStages {
				jenkinsTask := &devops.CICDTask{
					DomainEntity: domainlayer.DomainEntity{
						Id: fmt.Sprintf("%s:%s:%d:%s", "jenkins", "JenkinsTask", jenkinsBuild.ConnectionId,
							jenkinsBuild.DisplayName),
					},
					Name:         jenkinsBuild.DisplayName,
					Result:       jenkinsPipelineResult,
					Status:       jenkinsPipelineStatus,
					Type:         "CI/CD",
					DurationSec:  uint64(durationSec),
					StartedDate:  jenkinsBuild.StartTime,
					FinishedDate: jenkinsPipelineFinishedDate,
				}
				if jenkinsBuild.TriggeredBy != "" {
					jenkinsTask.PipelineId = fmt.Sprintf("%s:%s:%d:%s", "jenkins", "JenkinsPipeline",
						jenkinsBuild.ConnectionId, jenkinsBuild.TriggeredBy)
				} else {
					jenkinsTask.PipelineId = fmt.Sprintf("%s:%s:%d:%s", "jenkins", "JenkinsPipeline",
						jenkinsBuild.ConnectionId, jenkinsBuild.DisplayName)
				}
				jenkinsTask.RawDataOrigin = jenkinsBuild.RawDataOrigin
				results = append(results, jenkinsTask)

			}

			return results, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}
