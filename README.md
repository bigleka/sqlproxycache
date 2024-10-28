# sqlproxycache

##Objetivo 
A ideia principal é funcionar como um "man-in-the-midle" entre a aplicação e o banco de dados SQL Server fazendo cache dos resultados no Redis e consumindo esses resultados caso a mesma query seja executada novamente.

##Necessidade 
A principal necessidade é apenas ajudar aplicações mais antigas onde não temos como alterar o fonte original ou apenas falta vontade do desenvolvedor em usar um serviço de cache como Redis ou DragonFly. Não é substituir nenhum proxy já existente, mas ficar antes deles e a aplicação passa por ele antes de ir para o banco.

##Estado atual 
Ainda é muito embrionário, nem alpha direito. No arquivo do projeto ainda estou com o endereço do Cache configurado manualmente. A string de conexão ainda está "chumbada" no arquivo também. A aplicação manda um post para conseguir ter acesso ou manda um payload com a consulta e recebe o resultado Esse resultado é armazenado no Cache com a chave com um Hash baseado no Hash da query. Durante 5 minutos qualquer novo payload que tenha o mesmo Hash vai pegar do Cache.

##Próximas tentativas 
Usar um arquivo de configuração para o endereço do Cache e Banco Fazer uma passagem transparente da autenticação do banco, usuário e senha para o postgres um controle mais detalhado do TTL das chaves
