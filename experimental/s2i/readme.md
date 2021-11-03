# source-to-image experiments

```
oc create -f https://raw.githubusercontent.com/openshift/pipelines-tutorial/release-tech-preview-3/01_pipeline/01_apply_manifest_task.yaml
oc create -f https://raw.githubusercontent.com/openshift/pipelines-tutorial/release-tech-preview-3/01_pipeline/02_update_deployment_task.yaml
oc create -f https://raw.githubusercontent.com/tdolby-at-uk-ibm-com/ace-docker/master/experimental/s2i/pipeline/ace-demo-pipeline.yaml
oc create -f https://raw.githubusercontent.com/tdolby-at-uk-ibm-com/ace-docker/master/experimental/s2i/pipeline/s2i-ace-12-task.yaml
oc create -f https://raw.githubusercontent.com/tdolby-at-uk-ibm-com/ace-docker/master/experimental/s2i/pipeline/ace-pipeline-claim-pvc.yaml
```