# etutor
Golang tutorial desktop application.

To use it you need:
 1) The External packet (Golang GUI framework):  https://github.com/alkresin/external
 2) The GuiServer executable, which may be compiled from sources, hosted in https://github.com/alkresin/guiserver, or downloaded from http://www.kresin.ru/en/guisrv.html or a releases page of Github's repository.

<p align="center" markdown="1">
  <img src="screenshot/etutor_2.png" />
</p>

It is possible to change some options, editing the etutor.ini:
 - main font ( 'fontmain' tag );
 - editor window font ( 'fontcode' );
 - results window font ( 'fontresult' );
 - results window colors ( 'results' );
 - add/remove keywords to highlite ( 'hilighter' );
 - change the highlite scheme ( 'hiliopt' ), an alternative one is 'hiliopt_far', you may rename it to 'hiliopt' and the current 'hiliopt' to something else;
 - add new tutors, using the 'book' tag.

It is not necessary to keep the code in an xml file. You may use 'path' instead of 'code' with a path to your *.go file.

<b> Attention! Since October 6, 2023, we have been forced to use two-factor identification in order to 
   log in to github.com under your account. I can still do <i>git push</i> from the command line, but I can't
   use other services, for example, to answer questions. That's why I'm opening new projects on 
   https://gitflic.ru /, Sourceforge, or somewhere else. Follow the news on my website http://www.kresin.ru/

   Внимание! С 6 октября 2023 года нас вынуждили использовать двухфакторную идентификацию для того, чтобы 
   входить на github.com под своим аккаунтом. Я пока могу делать <i>git push<i> из командной строки, но не могу
   использовать другие сервисы, например, отвечать на вопросы. Поэтому новые проекты я открываю на 
   https://gitflic.ru/, Sourceforge, или где-то еще. Следите за новостями на моем сайте http://www.kresin.ru/ </b>
