$(function () {
  $('button').on('click', () => {
    $('form').slideToggle(alertFunc);
  });
  function alertFunc(){
    if ($(this).css('display') != 'none') {
      $("#btn").text("▲ 閉じる");
      $("#btn").css('color', 'orange')
      $("#btn").css('background', 'white')
    }else{
      $("#btn").text("新しいアルバムを作成する");
      $("#btn").css('color', '#fff')
      $("#btn").css('background', 'linear-gradient(45deg, #FFC107 0%, #ff8b5f 100%)')
    }
  };

  $('main').inertiaScroll({
   parent: $("#wrap")
 });
});
