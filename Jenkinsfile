/*
 * Copyright (C) 2022 SailPoint Technologies, Inc.  All rights reserved.
 */
@Library('sailpoint/jenkins-release-utils')_

/**
 * Jenkins pipeline for building and uploading sp-cli docker image.
 */
pipeline {
	agent none

	options {
		// Aborts job if run time is over 24 hours
		timeout(time: 24, unit: 'HOURS')

		// Add timestamps to console output
		timestamps()

		// Don't allow concurrent builds to run
		disableConcurrentBuilds()

		// Keep builds for a year + 30 days.
		buildDiscarder(logRotator(daysToKeepStr: '395'))
	}

	triggers {
		// Poll for changes every 5 minutes.
		pollSCM('H/5 * * * *')
	}

	environment {
		// The scrum which owns this component
		JIRA_PROJECT = "PLTCONN"

		// The name of the build artifact to generate
		BUILD_NUMBER = "${env.BUILD_NUMBER}"

		// The maximum amount of time (in minutes) to wait for a build
		BUILD_TIMEOUT = 20

		// The maximum amount of time (in minutes) for tests to take before they are auto failed.
		TEST_TIMEOUT = 10

		// The maximum amount of time (in minutes) to wait for a deploy
		DEPLOY_TIMEOUT = 30

		// Which room to report successes & failures too.
		SLACK_CHANNEL = "#team-eng-platform-connectivity-jnk"

		// The branch releases can be cut from.
		RELEASE_BRANCH = "main"

		// The name of service being released
		SERVICE_NAME = "sp-cli"
	}

	stages {
		stage('Build and push sp-cli') {
			when {
				branch env.RELEASE_BRANCH
			}
			steps {
				echo "${env.SERVICE_NAME} release pipeline for ${env.BUILD_NUMBER} is starting."
				sendSlackNotification(
						env.SLACK_CHANNEL,
						"${env.SERVICE_NAME} service release pipeline for <${env.BUILD_URL}|${env.BUILD_NUMBER}> is starting.",
						utils.NOTIFY_START
				)
				script {
					node {
						label 'devaws'
						checkout scm

						echo "Starting build of ${env.SERVICE_NAME}"

						sh("make VERSION=${env.BUILD_NUMBER} docker/push")

						//Git Config
						sh "git config --global user.email jenkins@construct.identitysoon.com"
						sh "git config --global user.name Jenkins"

						// Create and push a git tag for build
						TAG_NAME= "jenkins/${env.SERVICE_NAME}/${env.BUILD_NUMBER}"
						sh "git tag -a -f -m 'Built by Pipeline' ${TAG_NAME}"
						sh "git push origin tag ${TAG_NAME}"

						
					}
				}
			}
		}
	}

	post {
		success {
			sendSlackNotification(
					env.SLACK_CHANNEL,
					"${env.SERVICE_NAME} release pipeline for <${env.BUILD_URL}|${env.BUILD_NUMBER}> was successful.",
					utils.NOTIFY_SUCCESS
			)
		}
		failure {
			sendSlackNotification(
					env.SLACK_CHANNEL,
					"${env.SERVICE_NAME} release pipeline for <${env.BUILD_URL}|${env.BUILD_NUMBER}> failed.",
					utils.NOTIFY_FAILURE
			)
		}
		aborted {
			sendSlackNotification(
					env.SLACK_CHANNEL,
					"${env.SERVICE_NAME} release pipeline for <${env.BUILD_URL}|${env.BUILD_NUMBER}> was aborted.",
					utils.NOTIFY_ABORTED
			)
		}
	}
}
