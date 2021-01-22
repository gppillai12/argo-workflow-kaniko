package cmd

import (
	"bufio"
	"bytes"
	"github.com/fatih/color"
	docker "github.com/fsouza/go-dockerclient"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"yala/pkg/command/runner"
	"yala/pkg/config"
	"yala/pkg/util"
)

const bundleDir string = "bundle"
const dockerImageDir string = "docker-images"
const dockerCmd string = "docker"

var bundleCmd = &cobra.Command{
	Use:   "bundle",
	Short: "Bundle all images as tar file.",
	Long:  "Bundle all images as tar file.",
	Run: func(cmd *cobra.Command, args []string) {

		clusterConfig, err := config.NewFromFile(globalOptions.ClusterConfigFile)
		if err != nil {
			log.WithError(err).WithField("filename", globalOptions.ClusterConfigFile).Fatal("Failed to create  config object from config file.")
		}
		var bundlePath string
		var bundleName string
		//Creates directory if doesn't exist
		if clusterConfig != nil {
			bundleVersion := clusterConfig.Version
			bundlePath = globalOptions.HomeDir + "/" + clusterConfig.ClusterName + "/" + bundleDir + "/" + dockerImageDir
			bundleName = bundleDir + "/" + dockerImageDir + "-" + bundleVersion + ".tar.gz"
			if clusterConfig.Docker.Bundle == nil {
				log.Warn("Bundle path not provided, using the default path!")
				log.Info("Using bundlePath ", bundlePath, " For creating the docker images bundle!")
			} else {
				if clusterConfig.Docker.Bundle.Path != "" {
					bundlePath = clusterConfig.Docker.Bundle.Path + "/" + dockerImageDir
				}
				if clusterConfig.Docker.Bundle.Name != "" {
					bundleName = clusterConfig.Docker.Bundle.Name
				}
			}
			if err := util.Mkdir(bundlePath, 0750); err != nil {
				log.Fatal("Error creating bundle directory ", err)
			}
		} else {
			log.Fatal("Nil clusterConfig provided!")
		}

		log.Info("Preparing Bundle.")
		var pullImageList []string
		pullObjectList := clusterConfig.Docker.Pull
		au := &docker.AuthConfiguration{
			Username: "",
			Password: "",
		}
		for i, pullObject := range pullObjectList {
			color.Yellow("****** PULLING FROM REGISTRY %s %s%s%s %s ", pullObject.Registry, strconv.Itoa(i+1), "/", strconv.Itoa(len(pullObjectList)), " ******")
			pullImageList = pullObject.Images
			if pullObject.Auth == nil {
				log.Info("Proceeding without docker credentials!")
			} else {
				au = &docker.AuthConfiguration{
					Username: pullObject.Auth.Username,
					Password: pullObject.Auth.Password,
				}
			}
			pullImages(au, pullObject, pullImageList)
		}

		var buildImageList, finalImageList, dockerfileList []string
		buildImageList = getBuildList(clusterConfig.Docker.Build)
		if clusterConfig.Docker.Build != nil {
			dockerfileList = clusterConfig.Docker.Build
			buildImages(dockerfileList)
			finalImageList = append(pullImageList, buildImageList...)
		} else {
			finalImageList = pullImageList
		}
		util.WriteToFile(finalImageList, bundlePath+"/"+dockerImgListFile)
		var imageTarList []string
		for _, imageId := range finalImageList {
			imageTarList = append(imageTarList, tarImage(imageId, bundlePath))
		}
		createBundle(imageTarList, clusterConfig, bundleName, bundleDir)
		color.Green("****** Successfully created the bundle! ******")
	},
}

func getBuildList(dockerfileDirList []string) []string {
	var imgId, file string
	var imgList []string
	for _, dockerDir := range dockerfileDirList {
		file, _ = filepath.Abs(dockerDir)
		imgId = dockerParser(file)
		imgList = append(imgList, imgId)
	}
	return imgList
}

func pullImages(auth *docker.AuthConfiguration, pullObject *config.Pull, images []string) {
	dockerClient, _ := docker.NewClientFromEnv()
	for _, imageId := range images {
		imageName, imageTag := util.SplitBeforeAfter(imageId, ":")
		dockerPull(auth, pullObject.Registry, imageName, imageTag, dockerClient)
	}
}

func buildImages(dockerfiles []string) {
	for _, dockerfile := range dockerfiles {
		file, err := filepath.Abs(dockerfile)
		if err != nil {
			log.Fatal("Absolute filepath error for dockerfile ", dockerfile, "\n", err)
		}
		dockerBuildWithCLI(file)
	}
}

func createBundle(imageTarList []string, clusterConfig *config.ClusterConfig, bundleTarName, bundleDirName string) {
	bundleTarPath := globalOptions.HomeDir + "/" + clusterConfig.ClusterName + "/"
	// Archive to single file
	out, err := os.Create(bundleTarPath + "/" + bundleTarName)
	if err != nil {
		log.Fatalln("Error creating bundle stream :", err)
	}
	defer out.Close()
	if err := util.CreateArchive(imageTarList, out, bundleDirName); err != nil {
		log.Fatal("Error creating final bundle ", err)
	}
}

func tarImage(imageId, bundlePath string) string {
	dockerClient, _ := docker.NewClientFromEnv()
	imageName, imageTag := util.SplitBeforeAfter(imageId, ":")
	if strings.Contains(imageName, "/") {
		ss := strings.Split(imageName, "/")
		imageName = ss[len(ss)-1]
	}
	imageTarName := bundlePath + "/" + imageName + "-" + imageTag + ".tar"
	f, err := os.Create(imageTarName)
	if err != nil {
		log.Fatal("Error creating tar files ", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	log.Info("Exporting docker image ", imageId)
	opts := docker.ExportImagesOptions{Names: []string{imageId}, OutputStream: w}
	if err := dockerClient.ExportImages(opts); err != nil {
		log.Fatal("Error exporting image ", err)
	}
	w.Flush()
	return imageTarName
}
func dockerPull(auth *docker.AuthConfiguration, registry, imageName, imageTag string, dockerClient *docker.Client) {
	pullOptions := &docker.PullImageOptions{
		Registry:   registry,
		Repository: imageName,
		Tag:        imageTag,
	}
	log.Info("Pulling docker Image ", imageName+":"+imageTag)
	if err := dockerClient.PullImage(*pullOptions, *auth); err != nil {
		log.Fatal("Error pulling image", err)
	}
}

func dockerBuildWithCLI(dockerfile string) {
	initialDir, _ := os.Getwd()
	dockerfileDir := strings.Replace(dockerfile, "Dockerfile", "", -1)
	logger := log.StandardLogger()

	if err := os.Chdir(dockerfileDir); err != nil {
		log.Fatal("error changing directory ", err)
	}
	newDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Error getting current directory info ", err)
	}
	cmdRunner, err := runner.New(logger.Writer(), logger.Writer(), dockerfileDir)
	cmdRunner.SetDirectory(newDir)
	if err != nil {
		log.Fatal("Error creating new runner", err)
	}

	prov, err := cmdRunner.BinaryExists(dockerCmd)
	if err != nil {
		log.Fatal("error at docker setup ", err)
	}
	if !prov {
		log.Fatal("please install Docker before running the command")
	}
	err = cmdRunner.Run(dockerCmd, "build", "-f", dockerfile, ".")
	if err != nil {
		log.Fatal(err, "docker build command failed")
	}
	if err := os.Chdir(initialDir); err != nil {
		log.Fatal("error changing to initial directory ", err)
	}
}

func dockerParser(fileName string) string {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("failed to open ", err)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var text []string
	for scanner.Scan() {
		text = append(text, scanner.Text())
	}
	file.Close()
	var imageId string
	for _, line := range text {
		if strings.Contains(line, "FROM") {
			imageId = strings.TrimSpace(strings.Trim(line, "FROM"))
		}
	}
	return imageId
}

func dockerBuild(dockerFile string, dockerClient *docker.Client) {
	log.Info("Building dockerfile ", dockerFile)
	var buf bytes.Buffer
	dockerAuth := &docker.AuthConfiguration{
		Username: "userName",
		Password: "pass",
	}
	buildOptions := &docker.BuildImageOptions{
		Auth:         *dockerAuth,
		Dockerfile:   "../dockerfiles/dex/Dockerfile",
		OutputStream: &buf,
		ContextDir:   "../dockerfiles/dex/",
	}
	if err := dockerClient.BuildImage(*buildOptions); err != nil {
		log.Fatal("Error Building image ", err)
	}
}
