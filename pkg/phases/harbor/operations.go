package harbor

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/flanksource/karina/pkg/platform"
	"github.com/pkg/errors"
	"golang.org/x/sync/semaphore"
)

func ReplicateAll(p *platform.Platform) error {
	client, err := NewClient(p)
	if err != nil {
		return err
	}

	p.Infof("Listing replication policies")
	replications, err := client.ListReplicationPolicies()
	if err != nil {
		return fmt.Errorf("replicateAll: failed to list replication policies: %v", err)
	}
	for _, r := range replications {
		p.Infof("Triggering replication of %s (%d)\n", r.Name, r.ID)
		req, err := client.TriggerReplication(r.ID)
		if err != nil {
			return fmt.Errorf("replicateAll: failed to trigger replication: %v", err)
		}
		p.Infof("%s %s: %s  pending: %d, success: %d, failed: %d\n", req.StartTime, req.Status, req.StatusText, req.InProgress, req.Succeed, req.Failed)
	}
	return nil
}

func UpdateSettings(p *platform.Platform) error {
	client, err := NewClient(p)
	if err != nil {
		return err
	}
	p.Infof("Platform: %v", p)
	p.Infof("Settings: %v", *p.Harbor.Settings)
	return client.UpdateSettings(*p.Harbor.Settings)
}

func ListProjects(p *platform.Platform) ([]Project, error) {
	client, err := NewClient(p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create harbor client")
	}

	projects, err := client.ListProjects()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list projects")
	}

	return projects, nil
}

func ListImagesWithTags(p *platform.Platform, concurrency int) ([]Tag, error) {
	client, err := NewClient(p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create harbor client")
	}

	images, err := ListImages(p, concurrency)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list images")
	}

	lock := semaphore.NewWeighted(int64(concurrency))

	var wg sync.WaitGroup
	wg.Add(len(images))

	allTags := []Tag{}
	mtx := &sync.Mutex{}

	for _, i := range images {
		p.Debugf("artifacts count for %s/%s: %d\n", i.ProjectName, i.Name, i.ArtifactCount)

		go func(project, image string) {
			lock.Acquire(context.Background(), 1)
			defer func() {
				lock.Release(1)
				wg.Done()
			}()
			tags, err := client.ListTags(project, image)
			if err != nil {
				p.Errorf("failed to list tags for image %s in project %s: %v", image, project, err)
			}
			p.Tracef("tags count for %s/%s: %d\n", project, image, len(tags))

			for i := range tags {
				tags[i].ProjectName = project
				tags[i].RepositoryName = image

				if strings.HasPrefix(tags[i].Digest, "sha256:") {
					tags[i].Digest = strings.TrimPrefix(tags[i].Digest, "sha256:")
				}
			}

			mtx.Lock()
			allTags = append(allTags, tags...)
			mtx.Unlock()
		}(i.ProjectName, i.Name)
	}

	wg.Wait()

	sort.SliceStable(allTags, func(i, j int) bool {
		if allTags[i].ProjectName == allTags[j].ProjectName {
			if allTags[i].RepositoryName == allTags[j].RepositoryName {
				return allTags[i].Digest < allTags[j].Digest
			}
			return allTags[i].RepositoryName < allTags[j].RepositoryName
		}
		return allTags[i].ProjectName < allTags[j].ProjectName
	})

	return allTags, nil
}

func ListImages(p *platform.Platform, concurrency int) ([]Image, error) {
	client, err := NewClient(p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create harbor client")
	}

	projects, err := client.ListProjects()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list projects")
	}

	lock := semaphore.NewWeighted(int64(concurrency))

	var wg sync.WaitGroup
	wg.Add(len(projects))

	allImages := []Image{}
	mtx := &sync.Mutex{}

	for _, project := range projects {
		p.Debugf("listing images for project %s", project.Name)

		go func(projectName string) {
			lock.Acquire(context.Background(), 1)
			defer func() {
				lock.Release(1)
				wg.Done()
			}()
			images, err := client.ListImages(projectName)
			if err != nil {
				p.Errorf("failed to list images in project %s: %v", project, err)
			}

			for i := range images {
				images[i].ProjectName = projectName

				parts := strings.SplitN(images[i].Name, "/", 2)
				images[i].Name = parts[1]
			}

			mtx.Lock()
			allImages = append(allImages, images...)
			mtx.Unlock()
		}(project.Name)
	}

	wg.Wait()

	sort.SliceStable(allImages, func(i, j int) bool {
		if allImages[i].ProjectName == allImages[j].ProjectName {
			return allImages[i].Name < allImages[j].Name
		}
		return allImages[i].ProjectName < allImages[j].ProjectName
	})

	return allImages, nil
}
