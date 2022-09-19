pipeline {
    agent any
    options {
        lock(quantity: 1, resource: 'container-build-lock', variable: 'CONTAINER_BUILD_LOCK')
        buildDiscarder(logRotator(numToKeepStr:'10'))
    }
    stages {
        stage('Build') {
            steps {
                script {
                    version = sh(label: 'next_version', returnStdout: true, script: 'git log -1 --pretty=%h').trim()
                    currentBuild.displayName = "robot ${version}"
                    currentBuild.description = "robot container image build"
                    result = sh(label: 'build', returnStatus: true, script: "./build.sh")
                    if (result == 1) {
                        currentBuild.result = 'FAILURE'
                        echo '[FAILURE] build failed'
                        return
                    }
                }
            }
        }
    }
}