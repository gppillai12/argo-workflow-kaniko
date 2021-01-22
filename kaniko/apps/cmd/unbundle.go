package cmd

import (
	"github.com/fatih/color"
	docker "github.com/fsouza/go-dockerclient"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strings"
	"yala/pkg/command/runner"
	"yala/pkg/config"
	"yala/pkg/util"
)

const tarCmd string = "tar"
const dockerImgListFile string = "docker-images.txt"

var unbundleCmd = &cobra.Command{
	Use:   "unbundle",
	Short: "unBundles docker archive file.",
	Long:  "unBundle docker archive file & loads into docker registry.",
	Run: func(cmd *cobra.Command, args []string) {
		var bundlePath string
		var bundleName string
		clusterConfig, _ := config.NewFromFile(globalOptions.ClusterConfigFile)
		bundleVersion := clusterConfig.Version
		bundlePath = globalOptions.HomeDir + "/" + clusterConfig.ClusterName + "/" + bundleDir + "/"
		bundleName = dockerImageDir + "-" + bundleVersion + ".tar.gz"
		if clusterConfig.Docker.Bundle == nil {
			log.Warn("Bundle path not provided, using the default path!")
			log.Info("Using bundlePath ", bundlePath, " for creating the docker images bundle!")
		} else {
			if clusterConfig.Docker.Bundle.Path != "" {
				bundlePath = clusterConfig.Docker.Bundle.Path + "/" + bundleDir + "/"
			}
			if clusterConfig.Docker.Bundle.Name != "" {
				bundleName = clusterConfig.Docker.Bundle.Name
			}
		}

		color.Yellow("****** Starting unBundle Operation ******")
		unbundleImages(bundlePath, bundleName)
		color.Green("****** Successfully unbundled & loaded images into local registry! ******")
		color.Yellow("Tagging images to registry %s", clusterConfig.Docker.Push.Registry)
		imageList := util.ReadFileLineByLine(bundlePath + "/" + dockerImageDir + "/" + dockerImgListFile)
		pushRegistry := clusterConfig.Docker.Push.Registry
		var destinationRepo string
		var pushImageList []string
		for _, imageId := range imageList {
			imageName, imageTag := util.SplitBeforeAfter(imageId, ":")
			if strings.Contains(imageName, "/") {
				ss := strings.Split(imageName, "/")
				imageName = ss[len(ss)-1]
			}
			destinationRepo = pushRegistry + "/" + imageName
			dockerTag(imageId, destinationRepo, imageTag)
			pushImageList = append(pushImageList, destinationRepo+":"+imageTag)
		}
		color.Yellow("Pushing images to registry %s", clusterConfig.Docker.Push.Registry)
		pushObject := clusterConfig.Docker.Push
		if pushObject.Auth == nil {
			log.Fatal("Docker auth credentials missing!")
		}
		au := &docker.AuthConfiguration{
			Username: pushObject.Auth.Username,
			Password: pushObject.Auth.Password,
		}
		pushImages(au, pushObject, pushImageList)
		color.Green("****** Successfully pushed the images to the Registry! ******")
	},
}

func unbundleImages(bundlePath, tarBundle string) {
	err := extractTar(bundlePath, tarBundle)
	if err != nil {
		log.Fatal("error extracting tar file!", err)
	}
	log.Info("Bundle successfully extracted!")
	items, err := ioutil.ReadDir(bundlePath + bundleDir)
	if err != nil {
		log.Fatal("error opening extracted bundle! ", err)
	}
	for _, item := range items {
		log.Info("Loading image: ", item.Name())
		if err := dockerLoad(item.Name()); err != nil {
			log.Fatal("Error loading image : ", item.Name())
		}
		log.Info("Successfully loaded image: ", item.Name())
	}
}

func extractTar(bundlePath, tarFile string) error {
	logger := log.StandardLogger()
	cmdRunner, err := runner.New(logger.Writer(), logger.Writer(), bundlePath)
	if err != nil {
		log.Fatal("error creating new runner", err)
	}
	if err := os.Chdir(bundlePath); err != nil {
		log.Fatal("error changing directory ", err)
	}
	newDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Error getting current directory info ", err)
	}
	cmdRunner.SetDirectory(newDir)
	err = cmdRunner.Run(tarCmd, "xzf", tarFile)
	if err != nil {
		log.Fatal("error executing tar extract ", err)
	}
	return nil
}

func pushImages(auth *docker.AuthConfiguration, pushObject *config.Push, images []string) {
	dockerClient, _ := docker.NewClientFromEnv()
	for _, imageId := range images {
		imageName, imageTag := util.SplitBeforeAfter(imageId, ":")
		dockerPush(auth, pushObject.Registry, imageName, imageTag, dockerClient)
	}
}

func dockerPush(auth *docker.AuthConfiguration, registry, imageName, imageTag string, dockerClient *docker.Client) {
	pushImageOptions := &docker.PushImageOptions{
		Registry: registry,
		Name:     imageName,
		Tag:      imageTag,
	}
	log.Info("Pushing docker Image ", imageName+":"+imageTag)
	if err := dockerClient.PushImage(*pushImageOptions, *auth); err != nil {
		log.Fatal("Error pushing image", err)
	}
}

func dockerTag(sourceTag, destinationRepo, destinationTag string) {
	dockerClient, _ := docker.NewClientFromEnv()
	opts := docker.TagImageOptions{
		Repo: destinationRepo,
		Tag:  destinationTag,
	}
	dockerClient.TagImage(sourceTag, opts)
	log.Info("tagged ", sourceTag, " to ", destinationRepo+":"+destinationTag)
}

func dockerLoad(imageFile string) error {
	dockerClient, _ := docker.NewClientFromEnv()
	f, err := os.Open(imageFile)
	if err != nil {
		panic(err)
	}
	opts := docker.LoadImageOptions{InputStream: f}
	return dockerClient.LoadImage(opts)
}
