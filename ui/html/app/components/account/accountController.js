app
.controller("accountController",["$scope","$rootScope","accountService","$timeout",
function($scope,$rootScope,accountService,$timeout){
    $scope.list = [
        {
            name: "change_password",
            fontAwesome: "fas fa-redo"
        },
        {
            name: "logout",
            fontAwesome: "fas fa-sign-out-alt"
        }
    ]
    $scope.logout = function(){
        accountService.logout();
    }
    $scope.change_password = function(name){
        $timeout(function(){
            $rootScope.$emit("account_action",$scope.list);
        },100);
    }
    $rootScope.$on("logout",function(){
        accountService.logout();
    })
    $scope.state = false;
}])