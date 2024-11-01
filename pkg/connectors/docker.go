package connectors

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/PxyUp/fitter/pkg/config"
	"github.com/PxyUp/fitter/pkg/limitter"
	"github.com/PxyUp/fitter/pkg/logger"
	"github.com/docker/docker/api/types/container"
	imageTypes "github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/uuid"
	"io"
	"time"
)

const (
	defaultImage = "docker.io/zenika/alpine-chrome"

	pullTimeout = 60 * time.Second
)

var (
	errPullTimeout = errors.New("pull image timeout")

	defaultDockerFlags = []string{
		"--no-sandbox",
		"--headless",
		"--proxy-auto-detect",
		"--temp-profile",
		"--incognito",
		"--disable-logging",
		"--disable-gpu",
	}
)

func getFromDocker(url string, cfg *config.DockerConfig, logger logger.Logger) ([]byte, error) {
	ctxB := context.Background()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logger.Errorw("unable to connect to docker server", "error", err.Error())
		return nil, err
	}

	image := defaultImage
	if cfg.Image != "" {
		image = cfg.Image
	}

	if !cfg.NoPull {
		outPull, errPull := cli.ImagePull(ctxB, image, imageTypes.PullOptions{})
		if errPull != nil {
			logger.Errorw("unable to pull docker container", "error", errPull.Error())
			return nil, err
		}
		defer outPull.Close()

		logger.Infow("starting pull image", "image", image)
		pTimeout := pullTimeout
		if cfg.PullTimeout > 0 {
			pTimeout = time.Second * time.Duration(cfg.PullTimeout)
		}
		pullFinish := make(chan struct{})

		go func(dataForRead io.Reader) {
			_, _ = io.ReadAll(dataForRead)
			close(pullFinish)
		}(outPull)

		select {
		case <-pullFinish:
			logger.Infow("image pulled", "image", image)
		case <-time.After(pTimeout):
			logger.Errorw("timeout image pulling", "image", image)
			return nil, errPullTimeout
		}
	}

	t := timeout
	if cfg.Timeout > 0 {
		t = time.Second * time.Duration(cfg.Timeout)
	}

	ctxT, cancel := context.WithTimeout(ctxB, t)
	defer cancel()

	var args []string

	if len(cfg.Flags) != 0 {
		args = append(args, cfg.Flags...)
	} else {
		args = append(args, defaultDockerFlags...)
	}

	if len(cfg.Flags) == 0 {
		if cfg.Wait > 0 {
			args = append(args, fmt.Sprintf("--virtual-time-budget=%d", (time.Duration(cfg.Wait)*time.Millisecond).Milliseconds()))
		}

		args = append(args, fmt.Sprintf("--timeout=%d", t.Milliseconds()), "--dump-dom", url)
	}

	var entryPoint []string
	if cfg.EntryPoint != "" {
		entryPoint = append(entryPoint, cfg.EntryPoint)
	}

	resp, err := cli.ContainerCreate(ctxT, &container.Config{
		Image:      image,
		Cmd:        args,
		Entrypoint: entryPoint,
	}, &container.HostConfig{}, nil, nil, uuid.New().String())
	if err != nil {
		logger.Errorw("unable to create docker container", "error", err.Error())
		return nil, err
	}

	defer func() {
		if !cfg.Purge {
			return
		}
		removeCtx, cancelRemoveFn := context.WithTimeout(context.Background(), timeout)
		defer cancelRemoveFn()

		errRemove := cli.ContainerRemove(removeCtx, resp.ID, container.RemoveOptions{
			Force: true,
		})
		if errRemove != nil {
			logger.Errorw("unable to remove docker container", "error", err.Error())
			return
		}
		logger.Infow("container removed", "id", resp.ID)
	}()

	if instanceLimit := limitter.DockerLimiter(); instanceLimit != nil {
		errInstance := instanceLimit.Acquire(ctx, 1)
		if errInstance != nil {
			logger.Errorw("unable to acquire docker limit semaphore", "url", url, "error", errInstance.Error())
			return nil, errInstance
		}
		defer instanceLimit.Release(1)
	}

	err = cli.ContainerStart(ctxT, resp.ID, container.StartOptions{})
	if err != nil {
		logger.Errorw("unable to start docker container", "error", err.Error())
		return nil, err
	}

	defer func() {
		stopCtx, cancelStopFn := context.WithTimeout(context.Background(), timeout)
		defer cancelStopFn()
		errStop := cli.ContainerStop(stopCtx, resp.ID, container.StopOptions{})
		if errStop != nil {
			logger.Errorw("unable to stop docker container", "error", err.Error())
			return
		}
		logger.Infow("container stopped", "id", resp.ID)
	}()

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case errWait := <-errCh:
		if errWait != nil {
			logger.Errorw("unable to get docker status", "error", errWait.Error())
			return nil, errWait
		}
	case <-statusCh:
	}

	data, err := cli.ContainerLogs(ctxT, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		logger.Errorw("unable to get docker container logs", "error", err.Error())
		return nil, err
	}
	var outb, errb bytes.Buffer

	_, err = stdcopy.StdCopy(&outb, &errb, data)
	if err != nil {
		logger.Errorw("unable to copy logs from docker container", "error", err.Error())
		return nil, err
	}

	if errb.Len() > 0 {
		logger.Errorw("error during docker running", "url", url, "error", errb.String())
	}

	return outb.Bytes(), nil
}
