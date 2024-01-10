// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Drone Non-Commercial License
// that can be found in the LICENSE file.

package canceler

import (
	"testing"

	"github.com/drone/drone/core"
	"github.com/drone/drone/mock"
	"github.com/go-chi/chi/v5"

	"github.com/golang/mock/gomock"
)

func TestCancelPending_IgnoreEvent(t *testing.T) {
	ignore := []string{
		core.EventCron,
		core.EventCustom,
		core.EventPromote,
		core.EventRollback,
		core.EventTag,
	}
	for _, event := range ignore {
		s := new(service)
		err := s.CancelPending(noContext, nil, &core.Build{Event: event})
		if err != nil {
			t.Errorf("Expect cancel skipped for event type %s", event)
		}
	}
}

func TestCancelRunning_IgnoreEvent(t *testing.T) {
	ignore := []string{
		core.EventCron,
		core.EventCustom,
		core.EventPromote,
		core.EventRollback,
		core.EventTag,
	}
	for _, event := range ignore {
		s := new(service)
		err := s.CancelPending(noContext, nil, &core.Build{Event: event})
		if err != nil {
			t.Errorf("Expect cancel skipped for event type %s", event)
		}
	}
}

func TestCancel(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockStages := []*core.Stage{
		{Status: core.StatusPassing},
		{
			Status: core.StatusPending,
			Steps: []*core.Step{
				{Status: core.StatusPassing},
				{Status: core.StatusPending},
			},
		},
	}

	mockBuildCopy := new(core.Build)
	*mockBuildCopy = *mockBuild

	repos := mock.NewMockRepositoryStore(controller)

	events := mock.NewMockPubsub(controller)
	events.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil)

	builds := mock.NewMockBuildStore(controller)
	builds.EXPECT().Update(gomock.Any(), mockBuildCopy).Return(nil)

	users := mock.NewMockUserStore(controller)
	users.EXPECT().Find(gomock.Any(), mockRepo.UserID).Return(mockUser, nil)

	stages := mock.NewMockStageStore(controller)
	stages.EXPECT().ListSteps(gomock.Any(), mockBuild.ID).Return(mockStages, nil)
	stages.EXPECT().Update(gomock.Any(), mockStages[1]).Return(nil)

	steps := mock.NewMockStepStore(controller)
	steps.EXPECT().Update(gomock.Any(), mockStages[1].Steps[1]).Return(nil)

	status := mock.NewMockStatusService(controller)
	status.EXPECT().Send(gomock.Any(), mockUser, gomock.Any()).Return(nil)

	webhook := mock.NewMockWebhookSender(controller)
	webhook.EXPECT().Send(gomock.Any(), gomock.Any()).Return(nil)

	scheduler := mock.NewMockScheduler(controller)
	scheduler.EXPECT().Cancel(gomock.Any(), mockBuild.ID).Return(nil)

	c := new(chi.Context)
	c.URLParams.Add("owner", "octocat")
	c.URLParams.Add("name", "hello-world")
	c.URLParams.Add("number", "1")

	s := New(builds, events, repos, scheduler, stages, status, steps, users, webhook)
	err := s.Cancel(noContext, mockRepo, mockBuildCopy)
	if err != nil {
		t.Error(err)
	}
}

var (
	mockRepo = &core.Repository{
		ID:        1,
		Namespace: "octocat",
		Name:      "hello-world",
		Slug:      "octocat/hello-world",
		Counter:   42,
		Branch:    "master",
	}

	mockBuild = &core.Build{
		ID:           1,
		Number:       1,
		RepoID:       1,
		Status:       core.StatusPending,
		Event:        core.EventPush,
		Link:         "https://github.com/octocat/Hello-World/commit/7fd1a60b01f91b314f59955a4e4d4e80d8edf11d",
		Timestamp:    1299283200,
		Message:      "first commit",
		Before:       "553c2077f0edc3d5dc5d17262f6aa498e69d6f8e",
		After:        "7fd1a60b01f91b314f59955a4e4d4e80d8edf11d",
		Ref:          "refs/heads/master",
		Source:       "master",
		Target:       "master",
		Author:       "octocat",
		AuthorName:   "The Octocat",
		AuthorEmail:  "octocat@hello-world.com",
		AuthorAvatar: "https://avatars3.githubusercontent.com/u/583231",
		Sender:       "octocat",
	}

	mockUser = &core.User{
		ID:    1,
		Login: "octocat",
	}
)
