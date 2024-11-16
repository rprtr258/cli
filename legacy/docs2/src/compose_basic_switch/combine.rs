use bpaf::*;

pub fn options() -> OptionParser<bool> {
    let simple = short('s').long("simple").Switch();
    simple.to_options()
}
