app.config(['$routeProvider', '$locationProvider',
  function($routeProvider, $locationProvider) {
    if(window.location.pathname == "/dashboard.html"){
        $routeProvider.when('/', {
            templateUrl: 'app/components/dashboard/dashboard.html',
            controller: 'dashboardController'
        }).when('/:service', {
            templateUrl: 'app/components/dashboard/dashboard.html',
            controller: 'dashboardController'
        })
    }else if(window.location.pathname == "/index.html"){
        $routeProvider.when('/', {
            templateUrl: 'app/components/login/login.html',
            controller: 'loginController'
        })
    }else if(window.location.pathname == "/admin.html"){
        $routeProvider.when('/', {
            templateUrl: 'app/components/admin-dashboard/adminDashboard.html',
            controller: 'adminDashboardController'
        })
        .when('/:service', {
            templateUrl: 'app/components/admin-dashboard/adminDashboard.html',
            controller: 'adminDashboardController'
        });
    }else if (window.location.pathname == "/"){
        if(window.sessionStorage.getItem("is_admin") == "true"){
            window.location.href = "admin.html#!/get_services";
        }else if(window.sessionStorage.getItem("is_admin") == "false"){
            window.location.href = "/dashboard.html#!/all_jobs";
        }else{
            window.location.href = "/index.html";
        }
    }
        
    $locationProvider.html5Mode(false);
}])