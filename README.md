# android-localize-helper
Library and program for export android resource strings to csv and import them back after filling localized messages

## About
This project was created when nothing was found for export string resources from Android Studio project for translating them. Contains small library for loading resource files process them, export to csv, import from csv and save to Android Studio project resources. Also there is possibility to build exec-module for exporting and importing strings from Android Studio project resources

## Features

- looking for existing lcales in the project
- export all the translatable values to csx-file (including id, default locale valu and values for all or selected locales)
- import translated values from csv
- possibility to add locale from csv

## Getting Started

go get github.com/vc2402/localizer

go build github.com/vc2402/localizer