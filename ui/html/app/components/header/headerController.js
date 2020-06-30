app.controller("headerController",["$scope","$translate",function($scope,$translate){
    $scope.language = window.localStorage.getItem("lang");
    $scope.header = {};
    $scope.home = window.location.hash;
    $scope.language_list = [
        {
            "name": "English",
            "value": "en"
        },
        {
            "name": "Greek",
            "value": "el"
        }
    ]
    $scope.list = [
        {
            "name":"dashboard"
        }
    ]
    $scope.changeLanguage = function(lang) {
        $translate.use(lang).then(function(data){
            window.localStorage.setItem("lang",lang);
        }); 
    }
    $scope.item_clicked = function (item){
        $scope.list.forEach(function(value){
            $scope.header[value.name] = value.name == item;
        })
    }
    var route = window.location.hash.split("/");
    $scope.list.forEach(function(value){
        $scope.header[value.name] = route[1] == value.name;
    })
    // window.translator.translate();
}]);