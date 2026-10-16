package templates

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
)

// LegacyWorkflows are the CI workflows earlier forgego releases generated, which .github/workflows/.index.yml
// replaces, as SHA-256 hashes of their normalised text. Made by rendering the workflow at every release tag
// from v1.0.0 to v1.6.1, and at main before the split pipelines, in each mode (backend, library, website), with
// and without --docker, --no-testing, --no-linting and --no-versioning. A file matching one is forgego's,
// unchanged, and safe to remove.
var LegacyWorkflows = map[string]map[string]bool{
	".github/workflows/index.yml": {
		"03606faa52a36907f086fa06ed0f2c8dab484e13954d31037cec2ea6c5847f00": true,
		"05530910037b489a3a6e17f9405c7f5e6602b7711eb3fcc550865186e229d6c3": true,
		"05d52857fcb3f0d646d4a601c769e5e7d71aa0a08623de299ea64a35b29bf28e": true,
		"0766df525bc55b0193d941dee6102334b4ab3d8b85b6f1f0d79b24c59fe52ae0": true,
		"08b9bff915588d9af153646226bc309fee3a4d644268b75298730b37f4978d50": true,
		"08d529bfb034c88cb90ce75ce4980119e8909bde4a299a5c2dc6b130f16a3bc6": true,
		"0a146f6b55f15f25e714646fd5b2b5df165576d8fd62f0919c409d7827301889": true,
		"0abc32e4d0189d375262974c39ab8c6b0f976ab7e608790db0c4291253ebf0f8": true,
		"15e8b2ea70ac3607cd86fe4b44f2e762f861567eefee8840365170364c51a1cd": true,
		"16533c66585de678bc7da3f4529caa1fbeb14cd58f030c2b416285ca1c24fce2": true,
		"16fd7de7421e48977e1c7530281211626f1443ec4c5f93b17fb30bad61e1382e": true,
		"192360e108ceb8bc2695f57e9e7015dbdab29f95ae1b8ddf725c8d4279a35c88": true,
		"1e5f5b4e1e7d38ede2ddc9bbfa58b3c642a48c8f1e1d97a8f6e8cd2719493347": true,
		"26f4b441a22c02a9dd85f5ceccc31197207b647d673695b0d0cf91fb5471415d": true,
		"283bed79547994f19d43bce93432d4c2d49318ff3eae2b6835cc6d430e61dff5": true,
		"29f20f0c45790ebebefe4ce6d3c2d6291d0292c18b634374225828c8752cc78e": true,
		"3391f217063e7c121392c1a5356578cc779ef43d2f7a1cc3b30ae5ab1d1dbcda": true,
		"35d1b660849e881282ee442b80781059f5616d3af0f01d4ffdc4049a61a7ef0c": true,
		"3be3f4bf12e5cd18b4a4bc831d107350c3717c8873139a9d2b5445a9ba3667bc": true,
		"3c57a42f36735c51adeb7ca2b9ee0420d428de8b96d226d6b626d17c362f99b5": true,
		"3dbfa83ea74ac9ec6fb6846ce22f560190ef9bd4f36733c0d956dc9c73e93fe9": true,
		"480714768889f956a8627927e12ccd8b4c54aa305b162194aa64981b5cda8c0b": true,
		"4916f8b4a627fb88b11d8d36816f47010810c24fb573cac4e82d501a20e87e4c": true,
		"49be006a8e0d4c541e9e5742007dc36ad5bc529d98105e4c84b4f039401f2a6f": true,
		"4a102587c706b1107e6934b2bc523302f9ce0c608257c9f332952e4cf287a41c": true,
		"4c908c38d34df68db289aaebcf54097904897f964a4a75a468cfaffa08eb4713": true,
		"4d5e026eeee9da55f983e61b6368fb85c52d299b6c928fbf362bea1ea446deaa": true,
		"4dfbffe15567ffc3ccc828db7b7e9c93ff18cd6e5ae1f854735f2a5a57058543": true,
		"50b22e75fd0e30e5d80699ac340ddec128076937c84d2cc866898d6fcc799562": true,
		"50e8b05dcf0df55c3a61407f28c42b51348fe0ef0b2a0c1efb6e6f62b0272caf": true,
		"5715faac358fac587227461722ed1165e78711d0db46319d44122b5948fe8f8b": true,
		"5829c31988cb03959befc32719db304fd3a88a93b94f3d13deee5f0921b7ee13": true,
		"5b007c248f9924aacf88f9579bbc9e1ca219add8c8a1a1962ccbac7c11b3a46f": true,
		"5c07d4cab7b96c573c8cb770d326f0d10ed22135a2d4d656ea371f59531f2b0f": true,
		"5c8b7c088533e862a931c1a10fc571633a9e9acd43d49fbba133f91d2f5d300e": true,
		"6331aab327a0319e402a5fc6ed33f6a4fb8c21f0f5ad2f5daf2544ccc5d379a7": true,
		"65e6dadcb898204b78b60cd0e9bbee52dbe5987d5855aa85b9fbe3dbf7d29278": true,
		"68593cc774b2f30d0a7e25f061b489c6870565b4251ae219d643a7e6fcab243c": true,
		"68ae7c8db81ae81430e1fb16be69741a140eec92e07a6bcf7cbdddb05d524e30": true,
		"6905b63bc5425eb5592b7d6d74b15b8b5590e787e30b103226630c7af1e560c3": true,
		"6edc6afec9fcf2d5499c364f86da3db6e5b8f94bd71dc72177de1a938a173046": true,
		"716c8636f7aa14afa8421156d7422246ce5e70a440e886dc34c27ebacf787b44": true,
		"71fe6424070c75ffdc8a6babc1852384ba7d31e388fffdffca5b70af5287c504": true,
		"76ae13ef0dd5ab16e6069d84217582f8803b6b48c00794028e19bc0cb7291c76": true,
		"7864b6bd02f52427c4258a681ba7dc1c51b8f2bf8acf688811d1750229e4aa1d": true,
		"7a9a846e19e88d1561b4d749149a3b3ce4d921362ef9bd7b9de8b6b38a04de66": true,
		"7b8a047beaf2f87d23f11bb16cac0356f4da574a401030f174b8ab32d9f62f9f": true,
		"7e1c133d177b19744f3f3e01d12fcb536100f2fe46c252a85568ab2b9e94c690": true,
		"7f16bca95de3a036425d844ea4fd141eac8c6cd2609c429a8d17d7cc527cf566": true,
		"824db23e8e99658933032c5ca2d0f08ee66beb5670ef20344c3991536b343d61": true,
		"875da48abf07555b088b667f1f46c59703668af76b9ad20fad8fe708f4d50821": true,
		"87ed92918bb28c433f2c2509411ff0f078a3a02edcffa332425139813378a146": true,
		"898fd9fe54ea6eac00e6565344e3f38886c6761cf75d971a1b479cb49438c003": true,
		"8d73d2b4971c73925fa0e7ec568e2bfb5e64efe6681e1b941315364f057d7ecb": true,
		"90a642ce83cab1d7b9e342fbd7c67e8aca1b850997abf222c25e0e87b9435c1f": true,
		"9486f8632bfc305681e5349e1132d61d81158c076767678bad0c50d8e0e03d42": true,
		"95fd1fa2d2e68c26d3bb973fa2d8c11747b897adb4fdca1e34512f10cbbc800b": true,
		"99aeb694e137c307699add6b4dc58693a92665587be9927f6898026f378fe45a": true,
		"9c7bb6d178a0640c64a2678173b70bb84119ccbcd69c6dcdee788e97d5e1d8bb": true,
		"9cfb3cabf17bdc88c50a2294573b4b51a3a1bea0e51619944db80e1db111f3a5": true,
		"9d4b7946082165136ce9311353db580cd843f966f655fe1fda71f25cff4d71da": true,
		"9fb9ab3e152cf25e4ad132623c63e81835e89f11c5bff7c14268128cb240cde6": true,
		"a2319d178965b7cbb5f1877483d46a10d74c5076a23c7fcaadb6114b2a2e9ed2": true,
		"a5484c5e229180ba54d2660fb6c98d632478c43a756efe88737c673c35a41bee": true,
		"a5565eec3eaf5810f8b7563b2c0ab33934ce1f9e76decdf61005a49633e0c966": true,
		"a805372584184f46eae42bcaaae4b3f41cec40ee69b4d4091629f7be5d41fabf": true,
		"a97bda58651fc13475fffa55bad953532a33b409fb09e8ee65c5be87bc9f943b": true,
		"ab8d0b6262fe9323db65baa74a75a1779d46ffc411e9b5e7a4ba36c566a38086": true,
		"ad285983da17e0a881e3b28567151e62b3c005c70803788ce2d22cade39a0057": true,
		"adb3a6719f602c40dbfe62cbbd9ddc4817ca226ac603e9fbecc5110004255e04": true,
		"ae093ba2d830521f54ef83a7c161aee741bc83db57d77b0a2764e203a1f77d71": true,
		"af69b80232a918fd8c97f69d44779bb2cb5c7bc8c47b8e2a27115fe95610a3ad": true,
		"b019594394f146fc68b0f1273bee45c3048c9a8737768eba16797e0b597dfef6": true,
		"b1a5ef83467dc742e20e7389a6aacc53fa183a3c6faa64ea0ae319622669a7a3": true,
		"b2c88276caac4a30f3e6c31d9166a839d2618a257af1e912c489326c0e2b315b": true,
		"b49a94f9f76e15fb85fb6dc8f2166bf0d10afcfe597befbb2cb243fa9f3ed921": true,
		"b8933ae4a9691408dd6744df22c0d87b4902197846597ca753e945f8af32d342": true,
		"bbe44a6f1d0a67461eadab2c87013f7cb3097de4cdbf5a7da2234c5fccca2472": true,
		"bc80facefe3ad072fd24e0024885d64261fbb801f7620c39d5b5e262f4307f9c": true,
		"be0bef53c5bad505dd58170363240ff9a1834e0dc7fea71add786bec73e1495b": true,
		"c115dc0030703a71c82829172d91f6f4fbe154df4749b1a2a1fda1add05c357c": true,
		"c284ab91a27c610ad0442a7a8eca1f4de9f82dc1b0b9d00c1a2a968d0d1cabec": true,
		"c61c6125f3dec4fc6a7deb721284ae721c256fde68262617edb6b34a1636bdc8": true,
		"c7b0b17e57bac5c4a37e31eeedb00b8750fc4fc991592c227acd7cb45331ab01": true,
		"c99ff27be0dc115361a1e3e1bde530815063ff2b84315656f1cab1f6a6b96711": true,
		"cca300bc988b49b29bc5e479f2ecd139ef5c88f86d9172f1a98261412011abe3": true,
		"d10db27beb41e697642b1a13ec86ca3ff1b996b14d445a8064c937d43dfb8e24": true,
		"d24adbda292c49d6b4f213e7703d47425e20dbfc0994d32cdb42d8284836aa76": true,
		"d2c923b1f4c2d754669c57d32ce11b0bd600abcb0cb6eda21ca7be42164a0536": true,
		"d2e4d9cb0e1c453cc21b3290169876d524cb1b56fb030eeb83799f64cc05e51c": true,
		"d6cc57c38364e782472e60f88236982edefd8f0ecaf5fb864de219c85dfd1998": true,
		"d7d9e5bddf9eac5c1a91a54297732fdbbf1415e7136f0690d46eeec37e4dc182": true,
		"dd291b1c48bb674f894a9acdbc403523ca4651710ca24974c85095679a00e346": true,
		"dd58b8ddfe5c6d41f73cdffe3ce8b7009c291a0989984e98fe62374ff7a63c9f": true,
		"de87d7f6fc7d294afb838ad09a0f834cd377f90f24664925d3515968df7af24c": true,
		"deee3cbc80ae9e522078ca7268d28ec1c7c73f851162c69ba35a7f694459b569": true,
		"e10ccd818414b1b5053e2226505b08e763481d1cf9bc99a48d220c6f782f1498": true,
		"e1b07c862c5071fa8824b9cbb82b0371e149a2cc532c629465dd3cb19eca69a0": true,
		"e224941b709fdb22e51f9d4a75738bf8fac928592a45e0feba664c95cda03bfe": true,
		"e99d8e100da038f4b1c0190fea5c9acd3e61b662caff8b423aa5bd2750527501": true,
		"ef519ffdee7b3afb0aba0e87066d09719f24ab21b7d4e033fbb92a4e0352b44b": true,
		"f0683169ec4383a17a1b9bd2665774b684f491556be51998d7e114ad82c38120": true,
		"f2d684e1edd9d92d98ff7806637594eae6c6985d93e7592c034425b6ce9efebe": true,
		"f61dc5566238e092bd1bf4d7c41923247d7598ae34b5bc27560c48a53f055a82": true,
		"f65ef0f40a69af5f43a298d5608bf8a06a1af6fbf413c0cf7449669ea7f55ae4": true,
		"f6ae500b132e4d5fb0f4f300f81ed58dbbdb169048391ac9ce7f239fd2aed74b": true,
		"f921fb3b3ebc4b9d43efeec6e753cbb0592f9f2def0c229b392157f6c0b43db9": true,
		"fe4d3c0fbc676d9ac3b4898e4d272341c3160ca1772dbb634ace8cf01492ac6b": true,
	},
}

// Releases up to v1.3.0 wrote the project's Go version, which came from its go.mod or --go, into the workflow.
var legacyGoVersion = regexp.MustCompile(`(?m)^(\s*go_version: )'[^'\n]*'$`)

// NormaliseWorkflow drops what doesn't make a workflow the project's own: line endings, trailing spaces, blank
// lines, and the Go version an earlier release copied from go.mod.
func NormaliseWorkflow(content string) string {
	var lines []string
	for line := range strings.SplitSeq(strings.ReplaceAll(content, "\r\n", "\n"), "\n") {
		line = strings.TrimRight(line, " \t")
		if strings.TrimSpace(line) != "" {
			lines = append(lines, line)
		}
	}
	return legacyGoVersion.ReplaceAllString(strings.Join(lines, "\n"), "${1}'__GO__'")
}

// IsLegacyWorkflow reports whether file is exactly as an earlier forgego generated it.
func IsLegacyWorkflow(file, content string) bool {
	sum := sha256.Sum256([]byte(NormaliseWorkflow(content)))
	return LegacyWorkflows[file][hex.EncodeToString(sum[:])]
}
