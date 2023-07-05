// Global application script will go here.

function loadIntoModal (method, url, data) {
    $.ajax({
        url: url,
        type: method,
        data: jQuery.param(data) ,
        contentType: 'application/x-www-form-urlencoded; charset=UTF-8',
    }).done(function (response) {
        console.log(response)
        showModalContent(response);
    }).fail(function (xhr, status, err) {
        alert("Error: "+ status);
        console.log(status)
        console.log(err)
        console.log(xhr)
    })
}

function showModalContent (response) {
    $('#modalTitle').html(response.modalTitle);
    $('#modalContent').html(response.modalContent);
    $('#modal').modal('show');    
}

