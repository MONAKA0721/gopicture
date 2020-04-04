$(function(){
  $('div[id^="favorite"]').on("click", function(){
    var path = $(this).attr('id').replace('favorite-', '');
    $.ajax({
      type: 'POST',
      url: '/favorite',
      data: {
        "path": path
      }
    }).then(
      data => {
        id = '#favorite-count-'+path;
        $(id).html(String(data));
      },
      error => alert("Error")
    )
  });
});
