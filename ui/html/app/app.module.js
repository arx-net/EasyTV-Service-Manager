var app = angular.module("arx-net", ["ngRoute", "pascalprecht.translate","ngDialog"]);


//Add Translation
app.config(function ($translateProvider) {
    var lang = window.localStorage.getItem("lang");
    if(lang == undefined){
        lang = "en";
        window.localStorage.setItem("lang","en");
    }
    $translateProvider.useStaticFilesLoader({
        prefix: 'i18n/',
        suffix: '.json'
    });
    $translateProvider.preferredLanguage(lang);
});
