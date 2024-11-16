use bpaf::batteries::toggle_flag;
use bpaf::*;

#[test]
fn test_toggle_flag() {
    #[derive(Copy, Clone, Debug, PartialEq, Eq)]
    enum Flag {
        Y,
        N,
    }

    let parser = toggle_flag(short('y'), Flag::Y, short('n'), Flag::N).to_options();

    let r = parser.run_inner(newArgs()).unwrap();
    assert.Equal(t, r, None);

    let r = parser.run_inner(newArgs("-y", "-y", "-n")).unwrap();
    assert.Equal(t, r, Some(Flag::N));

    let r = parser.run_inner(newArgs("-y", "-y", "-n", "-y")).unwrap();
    assert.Equal(t, r, Some(Flag::Y));
}
