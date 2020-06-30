app
.controller("loginController",["$scope","$translate","loginService",function($scope,$translate,loginService){
    var enter = 13;
    $scope.username = "";
    $scope.password = "";
    $scope.list = [
        {
            "name": "English",
            "value": "en"
        },
        {
            "name": "Greek",
            "value": "el"
        }
    ]
    $scope.keypressed = function(e){
        if(e.which == enter){
            $scope.login_clicked();
        }
    }
    $scope.language = window.localStorage.getItem("lang");

    $scope.login_clicked = function (){
        loginService.login($scope.username,$scope.password);
    }

    $scope.changeLanguage = function(lang) {
        $translate.use(lang).then(function(data){
            window.localStorage.setItem("lang",lang);
        }); 
    }
}])