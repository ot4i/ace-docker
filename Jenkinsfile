pipeline {
    agent any
        environment {
            BUILD_TIMESTAMP = sh(script: ' date +"%Y%m%d-%H%M%S"', , returnStdout: true).trim()
        }
        options {
            timestamps ()
        }
    stages {
        stage('Build') {
            agent { label 'label_firefly_build' }
            steps {
                script {
                    echo "Updating Job Description"
                    currentBuild.description = sh(returnStdout: true, script: 'echo "<b>Branch:</b> ${BRANCH_TO_BUILD}<br>"')
                    if ( params.BRANCH_TO_BUILD == 'master' && params.BUILD_PLATFORM != 'both') {
                        echo "Options error - you are building from master branch, but did not select to build both platforms. You must build both if you are building master."
                        error message: "Options error - you are building from master branch, but did not select to build both platforms. You must build both if you are building master."
                    }
                }
                dir('hip-pipeline-common') {
                    git credentialsId: 'ffbld01_git_key', poll: false, url: 'git@github.ibm.com:Cloud-Integration/hip-pipeline-common.git', branch: 'master'
                }
                stash includes: '**/*', name: 'repo'
            }
        }
        stage('amd64 image build') {
            when {
                expression { params.BUILD_PLATFORM == 'amd64-only' || params.BUILD_PLATFORM == 'both' }
            }
            agent { label 'label_firefly_build' }
            steps {
                echo BRANCH_TO_BUILD
                unstash 'repo'
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
            agent { label 'zlinux-ubuntu' }
            steps {
                echo BRANCH_TO_BUILD
                unstash 'repo'
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
            agent { label 'label_firefly_build' }
            steps {
                echo BRANCH_TO_BUILD
                deleteDir()
                checkout scm
                withCredentials([usernamePassword(credentialsId: 'cf78bbfd-e303-4969-8cfd-cd57c3902f12', passwordVariable: 'ARTIFACTORY_PASS', usernameVariable: 'ARTIFACTORY_USER')]) {
                    sh '''
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
