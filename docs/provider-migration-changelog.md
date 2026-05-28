# Provider Migration Changelog

## 2026-05-28

- Corrected the MT example account root from `Assests:Test:Wx` to `Assets:Test:Wx`.
- The template runtime no longer treats `Assests:` as a valid account root. Keeping compatibility would preserve an invalid Beancount account spelling and make template validation less predictable.
- Replaced provider-shaped formatting helpers such as `.fixed8`, `.fixed3`, `.fixed2`, and `.pad6` with the generic `.format("...")` column method.
- Removed implicit securities trade merge behavior. The current HTSec split-row case is represented with ordinary rules: ignore the amount-only row, and compute the amount on the quantity/price row.
- Removed Chinese securities direction inference from the runtime.
- Removed special IR mapping actions such as `orderType`, `direction`, `txType`, `quantity`, `price`, `commission`, and the security/crypto account-unit action fields. HTSec, Huobi, and HXSec now express their old behavior with transaction headers, `vars`, metadata, tags, and explicit `postings`.
