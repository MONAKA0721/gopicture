$(function(){
  $('div[id^="favorite"]').on("click", function(){
    var albumHash = $(location).attr('pathname').replace('/show/', '');
    var fileName = $(this).attr('id').replace('favorite-', '');
    console.log(fileName);
    $.ajax({
      type: 'POST',
      url: '/favorite',
      data: {
        "albumHash": albumHash,
        "fileName": fileName
      }
    }).then(
      data => {
        id = '#favorite-count-' + fileName;
        $(id).html(String(data));
      },
      error => alert("Error")
    )
  });
});
