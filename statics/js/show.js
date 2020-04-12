$(function(){
  var favs = [];
  var favNum = [];
  var spans = [];
  $("span").each(function(e,v){
      spans.push(v);
    });
    for (let i=0; i < spans.length; i++){
      favNum.push(Number(spans[i].innerText));
    }
    favNum.sort(
      function(a,b){
        return (a < b ? 1 : -1);
      }
    );

    function GetNum1(){
      for (let i=0; i < spans.length; i++){
        if(favNum[0] == spans[i].innerText){
          return i;
        }
      }
    }
    function GetNum2(){
      for (let i=0; i < spans.length; i++){
        if(favNum[1] == spans[i].innerText){
          return i;
        }
      }
    }
    function GetNum3(){
      for (let i=0; i < spans.length; i++){
        if(favNum[2] == spans[i].innerText){
          return i;
        }
      }
    }
    function GetNum4(){
      for (let i=0; i < spans.length; i++){
        if(favNum[3] == spans[i].innerText){
          return i;
        }
      }
    }
    function GetNum5(){
      for (let i=0; i < spans.length; i++){
        if(favNum[4] == spans[i].innerText){
          return i;
        }
      }
    }

    if(spans[GetNum1()]){
      $('#' + `${spans[GetNum1()].id}`).parents('.picture').attr('style','grid-row:1/2;grid-column:1/4;');
      spans.splice(GetNum1(),1);
    }
    if(spans[GetNum2()]){
      $('#' + `${spans[GetNum2()].id}`).parents('.picture').attr('style','grid-row:2/3;grid-column:1/3;');
      spans.splice(GetNum2(),1);
    }
    if(spans[GetNum3()]){
      $('#' + `${spans[GetNum3()].id}`).parents('.picture').attr('style','grid-row:2/3;grid-column:3/4;');
      spans.splice(GetNum3(),1);
    }
    if(spans[GetNum4()]){
      $('#' + `${spans[GetNum4()].id}`).parents('.picture').attr('style','grid-row:3/4;grid-column:1/2;');
      spans.splice(GetNum4(),1);
    }
    if(spans[GetNum5()]){
      $('#' + `${spans[GetNum5()].id}`).parents('.picture').attr('style','grid-row:3/4;grid-column:2/4;');
      spans.splice(GetNum5(),1);
    }

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
    location.reload();
  });
});


function popupImage() {
  var popup = document.getElementById('js-popup');
  if(!popup) return;

  $('.js-show-popup').on("click", function(){
    var imgsrc = $(this).attr('src');
    $('.popup-inner').find('img').attr('src',imgsrc);
    popup.classList.toggle('is-show');
  });

  $('#js-close-btn').on("click", function(){
    popup.classList.toggle('is-show');
  });

  $('#js-black-bg').on("click", function(){
    popup.classList.toggle('is-show');
  });
}
popupImage();
