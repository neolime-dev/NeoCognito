# Como Contribuir com o NeoCognito

Primeiramente, obrigado pelo seu interesse em contribuir com o NeoCognito! Estamos felizes em ter você aqui. Toda contribuição é bem-vinda, desde a correção de um simples erro de digitação até a implementação de uma nova funcionalidade.

## Como Começar

1.  **Fork o Repositório:** Clique no botão "Fork" no canto superior direito da página do repositório no GitHub para criar uma cópia sua.
2.  **Clone seu Fork:** `git clone https://github.com/SEU-USUARIO/NeoCognito.git`
3.  **Crie uma Branch:** `git checkout -b minha-feature-incrivel` (Use um nome descritivo).
4.  **Faça suas Alterações:** Implemente sua funcionalidade ou correção de bug.
5.  **Commit suas Alterações:** `git commit -m "feat: Adiciona funcionalidade incrível"` (Veja nosso guia de mensagens de commit abaixo).
6.  **Envie para seu Fork:** `git push origin minha-feature-incrivel`
7.  **Abra um Pull Request:** No GitHub, vá para o repositório original e abra um "Pull Request" com uma descrição clara de suas alterações.

## Guia para Pull Requests

- **Mantenha a Simplicidade:** Prefira pull requests pequenos e focados. Não agrupe várias funcionalidades em um único PR.
- **Escreva uma Boa Descrição:** Explique o "quê" e o "porquê" de suas alterações. Se o seu PR resolve uma `Issue` existente, mencione o número dela (ex: `Resolve #123`).
- **Código Limpo:** Certifique-se de que seu código segue as convenções do Go. Rode `go fmt` e `go vet` antes de fazer o commit para garantir a qualidade.
- **Seja Paciente:** Faremos o nosso melhor para revisar seu PR o mais rápido possível.

## Padrão de Mensagens de Commit

Usamos um padrão para as mensagens de commit para manter o histórico limpo e organizado. Por favor, siga este formato:

`<tipo>: <assunto>`

**Tipos comuns:**
-   `feat`: Uma nova funcionalidade.
-   `fix`: Uma correção de bug.
-   `docs`: Alterações na documentação.
-   `style`: Alterações que não afetam o significado do código (espaços, formatação, etc).
-   `refactor`: Uma alteração de código que não corrige um bug nem adiciona uma funcionalidade.
-   `test`: Adicionando testes ou corrigindo testes existentes.
-   `chore`: Alterações em processos de build, ferramentas auxiliarias, etc.

**Exemplo:** `feat: Adiciona suporte para temas personalizados no config.toml`

Obrigado mais uma vez por sua contribuição!
