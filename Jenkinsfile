node {
    deleteDir()

    try {
        stage ('Checkout') {
        	checkout scm
        }

        if (env.BRANCH_NAME == 'master') {
            stage ('go build') {
				sh "mkdir bin"
				sh "./proto.sh"
				sh "GOOS=linux GOARCH=amd64 go build -o bin/api ./api"
				sh "GOOS=linux GOARCH=amd64 go build -o bin/storage ./storage"
				sh "GOOS=linux GOARCH=amd64 go build -o bin/processor ./processor"
				sh "GOOS=linux GOARCH=amd64 go build -o bin/model-parser ./model-parser"
            }

            stage ('Build Image') {
				sh "docker build --build-arg APP_NAME=api . -t hub.kube.tisserv.net/ugb-api:v${env.BUILD_NUMBER}"
				sh "docker build --build-arg APP_NAME=storage . -t hub.kube.tisserv.net/ugb-storage:v${env.BUILD_NUMBER}"
				sh "docker build --build-arg APP_NAME=processor . -t hub.kube.tisserv.net/ugb-processor:v${env.BUILD_NUMBER}"
				sh "docker build --build-arg APP_NAME=model-parser . -t hub.kube.tisserv.net/ugb-model-parser:v${env.BUILD_NUMBER}"
            }

            stage ('Push&Clean Image') {
                sh "docker push hub.kube.tisserv.net/ugb-api:v${env.BUILD_NUMBER}"
                sh "docker push hub.kube.tisserv.net/ugb-storage:v${env.BUILD_NUMBER}"
                sh "docker push hub.kube.tisserv.net/ugb-processor:v${env.BUILD_NUMBER}"
                sh "docker push hub.kube.tisserv.net/ugb-model-parser:v${env.BUILD_NUMBER}"

                sh "docker tag hub.kube.tisserv.net/ugb-api:v${env.BUILD_NUMBER} hub.kube.tisserv.net/ugb-api:latest"
                sh "docker tag hub.kube.tisserv.net/ugb-storage:v${env.BUILD_NUMBER} hub.kube.tisserv.net/ugb-storage:latest"
                sh "docker tag hub.kube.tisserv.net/ugb-processor:v${env.BUILD_NUMBER} hub.kube.tisserv.net/ugb-processor:latest"
                sh "docker tag hub.kube.tisserv.net/ugb-model-parser:v${env.BUILD_NUMBER} hub.kube.tisserv.net/ugb-model-parser:latest"

                sh "docker push hub.kube.tisserv.net/ugb-api:latest"
                sh "docker push hub.kube.tisserv.net/ugb-storage:latest"
                sh "docker push hub.kube.tisserv.net/ugb-processor:latest"
                sh "docker push hub.kube.tisserv.net/ugb-model-parser:latest"
            }

            stage ('deploy kube.tisserv.net') {
               sh '''
                   cd infra-cloud

                   terraform init
                   terraform validate .
                   terraform plan -var DOCKER_IMG_TAG=v${BUILD_NUMBER}
                   terraform refresh -var DOCKER_IMG_TAG=v${BUILD_NUMBER}
                   terraform apply -var DOCKER_IMG_TAG=v${BUILD_NUMBER} -auto-approve
               '''
            }

            stage ('deploy tisworkstation') {
               sh '''
                   cd deploy/backend
                   docker-compose up -d
               '''
            }
        }
    } catch (err) {
        throw err
    }
}
