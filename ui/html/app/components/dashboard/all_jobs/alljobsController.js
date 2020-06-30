app
.controller("alljobsController",["$scope","dashboardService","requestService","modalService","errorService",
function($scope,dashboardService,requestService,modalService,errorService){
    $scope.selected_service = {};
    
    $scope.get_jobs = function(){
        dashboardService.get_jobs()
        .then(function(response){
            if(response.data.code == 200){
                response.data.jobs = get_date_from_epoch(response.data.jobs);
                $scope.job_list = response.data.jobs;
                $(document).ready(function(){
                    $('#job_list').DataTable({
                        order: [0, 'desc'],
                        retrieve: true,
                    });                 
                })
            }else{
                var error = errorService.handle_error(response.data.code);
                modalService.open_popup(modalService.modals.popup,error);
            }
        }).catch(function(response){
            var error = errorService.handle_error(response.data.code);
            error.message = response.data.message;
            modalService.open_popup(modalService.modals.popup,error);
        });;
    }
    $scope.show_output = function(output){
        data = {
            title: "output",
            data: output,
            buttons: modalService.buttons["ok"]
        }
        modalService.open_popup(modalService.modals.output,data);
    }
    $scope.cancel_job = function(job) {
        var data = {
            param: job.id,
            title : 'cancel_job_title',
            message: 'cancel_job_message',
            buttons: modalService.buttons["yes/no"]
        }
   
        modalService.open_dialog(modalService.modals.dialog,data,(ok)=>{
            if(ok == 1){
                console.log("Cancel job " + job.id)
                dashboardService.cancel_job(job.id)
                .then(function(response) {
                    if (response.data.code == 200) {
                        $scope.get_jobs();
                    } else {
                        console.log("Failed to delete job=" +job.id +
                         " code=" + response.data.code +
                         " description=" + response.data.description)
                    }
                })
            }
        });
    }
    function get_date_from_epoch(array){
        array.forEach(elm => {
            if(elm.completion_date != null){
                var date  = new Date(elm.completion_date * 1000);
                elm.completion_date = date.getDate() + '-' + (date.getMonth() + 1)+ '-' + date.getFullYear();
            }
            if(elm.creation_date != null){
                var date  = new Date(elm.creation_date * 1000);
                elm.creation_date = date.getDate() + '-' + (date.getMonth() + 1)+ '-' + date.getFullYear();
            }
            var date  = new Date(elm.expiration_date * 1000);
            elm.expiration_date = date.getDate() + '-' + (date.getMonth() + 1)+ '-' + date.getFullYear();
            var date  = new Date(elm.publication_date * 1000);
            elm.publication_date = date.getDate() + '-' + (date.getMonth() + 1)+ '-' + date.getFullYear();
        });
        return array;
    }
    $(document).ready(function(){
        setTimeout(function(){
            $('.date').datepicker({
                dateFormat:"mm-dd-yy",
                minDate: 0
            });
        },100)
    });

    if($scope.leftbar.all_jobs == true){
        $scope.get_jobs();
    }
}])