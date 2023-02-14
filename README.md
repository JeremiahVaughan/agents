This repo produces images used to execute CICD operations. 

### Reason these agents are built in their own repository
Originally I would build agents as needed at the start of pipelines where they were needed. But this added an extra 20 minimum seconds even if there weren't any changes to the build agent. Since changes to the agents are not common it makes sense to extract them into their own pipeline. 