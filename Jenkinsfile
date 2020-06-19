pipeline {
    agent any
        environment {
            BUILD_TIMESTAMP = sh(script: ' date +"%Y%m%d-%H%M%S"', , returnStdout: true).trim()
        }
        options {
            timestamps ()
        }
    stages {
        stage('amd64 image build') {
            when {
                expression { params.BUILD_PLATFORM == 'amd64-only' || params.BUILD_PLATFORM == 'both' }
            }
            agent { label 'cf_slave' }
            steps {
                echo BRANCH_TO_BUILD
                deleteDir()
                checkout scm
                dir('ot4i-ace-docker') {
                    git credentialsId: 'ffbld01_git_key', poll: false, url: 'git@github.ibm.com:Cloud-Integration/ot4i-ace-docker.git', branch: "${params.BRANCH_TO_BUILD}"
                }
                dir('hip-pipeline-common') {
                    git credentialsId: 'ffbld01_git_key', poll: false, url: 'git@github.ibm.com:Cloud-Integration/hip-pipeline-common.git', branch: 'master'
                }
                dir('firefly-software-build-scripts') {
                    git credentialsId: 'ffbld01_git_key', poll: false, url: 'git@github.ibm.com:Cloud-Integration/firefly-software-build-scripts.git', branch: "${params.FIREFLY_SOFTWARE_BUILD_SCRIPTS_BRANCH}"
                }
                withCredentials([usernamePassword(credentialsId: 'cf78bbfd-e303-4969-8cfd-cd57c3902f12', passwordVariable: 'ARTIFACTORY_PASS', usernameVariable: 'ARTIFACTORY_USER'), usernamePassword(credentialsId: '37361c4d-f3f7-4bf4-97a0-48463a5d2091', passwordVariable: 'GITHUB_API_TOKEN', usernameVariable: 'GITHUB_API_USER'), usernamePassword(credentialsId: 'cf78bbfd-e303-4969-8cfd-cd57c3902f12', passwordVariable: 'NPM_PASS', usernameVariable: 'NPM_USER'), string(credentialsId: 'APPCONNECT_NPM_AUTH', variable: 'NPM_AUTH')]) {
                  sh '''
                    bash -c "
                    pwd
                    ls -l
                    ls -l ${WORKSPACE}/jenkins-build-scripts
                    ${WORKSPACE}/jenkins-build-scripts/jenkins-build-script.sh
                    "
                  '''
                }
            }
        }
        stage('s390x image build') {
            when {
                expression { params.BUILD_PLATFORM == 's390x-only' || params.BUILD_PLATFORM == 'both' }
            }
            agent { label 'zlinux-ACEcc' }
            steps {
                echo BRANCH_TO_BUILD
                deleteDir()
                checkout scm
                dir('ot4i-ace-docker') {
                    git credentialsId: 'ffbld01_git_key', poll: false, url: 'git@github.ibm.com:Cloud-Integration/ot4i-ace-docker.git', branch: "${params.BRANCH_TO_BUILD}"
                }
                dir('hip-pipeline-common') {
                    git credentialsId: 'ffbld01_git_key', poll: false, url: 'git@github.ibm.com:Cloud-Integration/hip-pipeline-common.git', branch: 'master'
                }
                dir('firefly-software-build-scripts') {
                    git credentialsId: 'ffbld01_git_key', poll: false, url: 'git@github.ibm.com:Cloud-Integration/firefly-software-build-scripts.git', branch: "${params.FIREFLY_SOFTWARE_BUILD_SCRIPTS_BRANCH}"
                }
                withCredentials([usernamePassword(credentialsId: 'cf78bbfd-e303-4969-8cfd-cd57c3902f12', passwordVariable: 'ARTIFACTORY_PASS', usernameVariable: 'ARTIFACTORY_USER'), usernamePassword(credentialsId: '37361c4d-f3f7-4bf4-97a0-48463a5d2091', passwordVariable: 'GITHUB_API_TOKEN', usernameVariable: 'GITHUB_API_USER'), usernamePassword(credentialsId: 'cf78bbfd-e303-4969-8cfd-cd57c3902f12', passwordVariable: 'NPM_PASS', usernameVariable: 'NPM_USER'), string(credentialsId: 'APPCONNECT_NPM_AUTH', variable: 'NPM_AUTH')]) {
                  sh '''
                    bash -c "
                    pwd
                    ls -l
                    ls -l ${WORKSPACE}/jenkins-build-scripts
                    ${WORKSPACE}/jenkins-build-scripts/jenkins-build-script.sh
                    "
                  '''
                }
            }
        }
        stage('Create Docker Manifests') {
            agent { label 'cf_slave' }
            steps {
                echo BRANCH_TO_BUILD
                deleteDir()
                checkout scm
                withCredentials([usernamePassword(credentialsId: 'cf78bbfd-e303-4969-8cfd-cd57c3902f12', passwordVariable: 'ARTIFACTORY_PASS', usernameVariable: 'ARTIFACTORY_USER')]) {
                    sh '''
                        # the docker manifest command is an experimental feature
                        # so need to enable the docker client side experimental feature
                        mkdir /home/jenkins/.docker
                        echo '
                        {
                            "experimental": "enabled"
                        }
                        ' > /home/jenkins/.docker/config.json
                        bash -c "
                        pwd
                        ls -la
                        ${WORKSPACE}/jenkins-build-scripts/create-docker-manifests.sh
                        "
                    '''
                }
            }
        }
        stage ('Trigger Image Promotion Build') {
            steps {
                script {
                        echo "Triggering image promotion build "
                        def ImageName = "ace-server"
                        def tag = "${TAG_VERSION}-${BUILD_TIMESTAMP}"
                        def updateJSON = "false"
                        if (env.BRANCH_TO_BUILD == 'master') {
                            echo "Master branch so updating JSON doc"
                            updateJSON = "true"
                        }
                        build job: 'ibm-appconnect-operator-test-images', wait: true, parameters: [
                            [$class: 'StringParameterValue', name: "IMAGE_NAME", value: ImageName],
                            [$class: 'StringParameterValue', name: "UPDATE_GOOD_IMAGES_JSON", value: updateJSON],
                            [$class: 'StringParameterValue', name: 'IMAGE_TAG', value: tag]
                        ]
                }
            }
        }
    }
    post {
        fixed {
            slackSend channel: '#appcon-monza-feed', message: '*' + JOB_NAME + '*\n Successfully built branch - ' + BRANCH_TO_BUILD + '\nSee - ' + BUILD_URL, color: 'good'
        }
        failure {
            slackSend channel: '#appcon-monza-feed', message: '*' + JOB_NAME + '*:\n Failed to build branch - ' + BRANCH_TO_BUILD + '\nSee - ' + BUILD_URL, color: '#AA0114'
        }
    }
}
